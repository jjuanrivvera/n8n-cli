// Package commands wires the cobra command tree: global flags, client
// construction, output rendering, and every resource and meta command.
package commands

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/auth"
	"github.com/jjuanrivvera/n8n-cli/internal/config"
	"github.com/jjuanrivvera/n8n-cli/internal/output"
)

// A single bufio.Reader is shared across all interactive prompts in a command run.
// Using a fresh bufio.Reader per prompt would over-read os.Stdin (bufio buffers up
// to 4KB), so a second prompt — or a piped `echo key | n8nctl init` — would lose
// input. The reader is rebuilt when os.Stdin changes (e.g. between tests).
var (
	stdinFile *os.File
	stdinRdr  *bufio.Reader
)

func stdinReader() *bufio.Reader {
	if stdinRdr == nil || stdinFile != os.Stdin {
		stdinFile = os.Stdin
		stdinRdr = bufio.NewReader(os.Stdin)
	}
	return stdinRdr
}

// Global persistent flags, bound in init().
var (
	flagProfile   string
	flagOutput    string
	flagBaseURL   string
	flagAPIKey    string
	flagRPS       float64
	flagDryRun    bool
	flagShowToken bool
	flagVerbose   bool
	flagNoColor   bool
	flagQuiet     bool
	flagColumns   []string
	flagNoHeader  bool
	flagJQ        string
)

var rootCmd = &cobra.Command{
	Use:   "n8nctl",
	Short: "Control any n8n instance from the terminal via its public API",
	Long: `n8nctl is a portable, single-binary client for the n8n public REST API.

It manages workflows, executions, credentials, tags, variables, projects and
users on any n8n instance — self-hosted or Cloud — over HTTPS with an API key.

Multi-instance is first class: define named profiles, store each instance's API
key in your OS keyring, and switch with --profile or "n8nctl config use <name>".`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       "", // set by version command / ldflags via SetVersionTemplate
	// Validate the output format up front, so an invalid -o fails immediately
	// rather than after a network round-trip.
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		_, err := outputFormat()
		return err
	},
}

// RootCmd exposes the root command (used by docs generation and tests).
func RootCmd() *cobra.Command { return rootCmd }

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagProfile, "profile", "", "config profile (instance) to use [env: N8NCTL_PROFILE]")
	pf.StringVarP(&flagOutput, "output", "o", "", "output format: table|json|yaml|csv|id [env: N8NCTL_OUTPUT]")
	pf.StringVar(&flagBaseURL, "base-url", "", "override the instance base URL (e.g. https://host/api/v1)")
	pf.StringVar(&flagAPIKey, "api-key", "", "override the API key (prefer keyring via 'auth login')")
	pf.Float64Var(&flagRPS, "rps", 0, "client-side rate limit in requests/sec (0 = use config/default)")
	pf.BoolVar(&flagDryRun, "dry-run", false, "print the equivalent curl and send no request")
	pf.BoolVar(&flagShowToken, "show-token", false, "do not redact the API key in --dry-run output")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose (debug) logging to stderr")
	pf.BoolVar(&flagNoColor, "no-color", false, "disable colored output [env: NO_COLOR]")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "suppress non-essential chatter")
	pf.StringSliceVar(&flagColumns, "columns", nil, "comma-separated columns for table/csv output")
	pf.BoolVar(&flagNoHeader, "no-header", false, "hide the table header row")
	pf.StringVar(&flagJQ, "jq", "", "apply a jq program to the result (e.g. '.[].id'); implies JSON input")
}

// Execute expands user aliases then runs the command tree with a cancellable
// context (wired to OS signals by main), so every command's cmd.Context() is
// cancelled on Ctrl-C.
func Execute(ctx context.Context) error {
	expandAliases()
	return rootCmd.ExecuteContext(ctx)
}

// --- configuration loading (resolved once per process) ---

var (
	cfgOnce sync.Once
	cfgVal  *config.Config
	cfgErr  error
)

func loadConfig() (*config.Config, error) {
	cfgOnce.Do(func() { cfgVal, cfgErr = config.Load() })
	return cfgVal, cfgErr
}

// activeProfile returns the resolved active profile name (flag > env > default).
func activeProfile() (string, *config.Config, error) {
	c, err := loadConfig()
	if err != nil {
		return "", nil, err
	}
	return c.ActiveProfileName(flagProfile), c, nil
}

// newLogger builds a stderr logger at the level implied by --verbose/config.
func newLogger(level string) *slog.Logger {
	lvl := slog.LevelWarn
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "error":
		lvl = slog.LevelError
	}
	if flagVerbose {
		lvl = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
}

// getAPIClient builds an *api.Client from the resolved profile, env, and flags.
// Precedence for each value is flag > env > config file > default.
func getAPIClient(cmd *cobra.Command) (*api.Client, error) {
	profileName, c, err := activeProfile()
	if err != nil {
		return nil, err
	}
	resolved := c.Resolve(profileName)

	baseURL := resolved.BaseURL
	if flagBaseURL != "" {
		baseURL = flagBaseURL
	}
	if baseURL == "" {
		return nil, fmt.Errorf("no base URL for profile %q — run `n8nctl init` or pass --base-url", profileName)
	}

	// API key precedence: flag > env (handled in Resolve) > keyring.
	apiKey := resolved.APIKey
	if flagAPIKey != "" {
		apiKey = flagAPIKey
	}
	if apiKey == "" {
		apiKey = auth.Lookup(profileName)
	}
	if apiKey == "" && !flagDryRun {
		return nil, fmt.Errorf("no API key for profile %q — run `n8nctl auth login` (or set N8NCTL_API_KEY)", profileName)
	}

	rps := resolved.RequestsPerSecond
	if flagRPS > 0 {
		rps = flagRPS
	}

	logger := newLogger(resolved.LogLevel)
	if !strings.HasPrefix(strings.ToLower(baseURL), "https://") && !flagQuiet {
		fmt.Fprintf(os.Stderr, "warning: base URL %q is not HTTPS; the API key will be sent in clear text\n", baseURL)
	}

	return api.New(
		api.WithBaseURL(baseURL),
		api.WithAPIKey(apiKey),
		api.WithLogger(logger),
		api.WithRequestsPerSecond(rps),
		api.WithDryRun(flagDryRun, flagShowToken, cmd.OutOrStdout()),
	), nil
}

// clientForProfile builds a client for a specific named profile (used by
// cross-instance features like sync, backup, and restore). Unlike getAPIClient
// it ignores --base-url/--api-key flags so it always targets the named instance.
// dryRun is passed explicitly so a "source" read can be live while a "destination"
// write is previewed.
func clientForProfile(cmd *cobra.Command, profileName string, dryRun bool) (*api.Client, error) {
	c, err := loadConfig()
	if err != nil {
		return nil, err
	}
	resolved := c.Resolve(profileName)
	if resolved.BaseURL == "" {
		return nil, fmt.Errorf("no base URL for profile %q — run `n8nctl init --profile %s`", profileName, profileName)
	}
	apiKey := resolved.APIKey
	if apiKey == "" {
		apiKey = auth.Lookup(profileName)
	}
	if apiKey == "" && !dryRun {
		return nil, fmt.Errorf("no API key for profile %q — run `n8nctl auth login --profile %s`", profileName, profileName)
	}
	return api.New(
		api.WithBaseURL(resolved.BaseURL),
		api.WithAPIKey(apiKey),
		api.WithLogger(newLogger(resolved.LogLevel)),
		api.WithRequestsPerSecond(resolved.RequestsPerSecond),
		api.WithDryRun(dryRun, flagShowToken, cmd.OutOrStdout()),
	), nil
}

// getReadClient returns a client that always performs reads, even under
// --dry-run. Commands that must read remote state to compute a plan or diff
// (apply, diff, lint --remote) use this; apply then skips its writes itself when
// --dry-run is set, printing the plan instead.
func getReadClient(cmd *cobra.Command) (*api.Client, error) {
	profile, _, err := activeProfile()
	if err != nil {
		return nil, err
	}
	return clientForProfile(cmd, profile, false)
}

// outputFormat resolves the output format: --output flag > config/env > table.
func outputFormat() (output.Format, error) {
	if flagOutput != "" {
		return output.Parse(flagOutput)
	}
	_, c, err := activeProfile()
	if err != nil {
		return output.Table, nil //nolint:nilerr // fall back to default on config error
	}
	return output.Parse(c.Resolve(c.ActiveProfileName(flagProfile)).OutputFormat)
}

// render writes data to stdout in the resolved format with column/color options.
// defaultCols supplies the table/csv columns when the user did not pass --columns
// (each resource declares a sensible set so list output stays readable instead of
// dumping every nested field).
func render(cmd *cobra.Command, data any, defaultCols ...string) error {
	// --jq short-circuits formatting: filter the data and emit JSON results.
	if flagJQ != "" {
		return output.ApplyJQ(cmd.OutOrStdout(), data, flagJQ)
	}
	format, err := outputFormat()
	if err != nil {
		return err
	}
	cols := flagColumns
	if len(cols) == 0 {
		cols = defaultCols
	}
	noColor := flagNoColor || os.Getenv("NO_COLOR") != ""
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	opts := output.Options{
		Columns:  cols,
		NoColor:  noColor,
		Color:    isTTY && !noColor,
		NoHeader: flagNoHeader,
	}
	return output.Render(cmd.OutOrStdout(), data, format, opts)
}

// dryRunNotice prints a hint after a dry-run mutation so users know nothing happened.
func dryRunNotice(cmd *cobra.Command) {
	if flagDryRun && !flagQuiet {
		fmt.Fprintln(cmd.ErrOrStderr(), "(dry-run: no request was sent)")
	}
}
