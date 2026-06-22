package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	n8ncli "github.com/jjuanrivvera/n8n-cli"
)

// agentSkillDirs maps an agent name to {global, project} skills directories,
// matching where `npx skills` installs. %h is expanded to the home directory.
var agentSkillDirs = map[string][2]string{
	"claude":   {"%h/.claude/skills", ".claude/skills"},
	"cursor":   {"%h/.cursor/skills", ".agents/skills"},
	"windsurf": {"%h/.codeium/windsurf/skills", ".windsurf/skills"},
	"codex":    {"%h/.codex/skills", ".agents/skills"},
	"gemini":   {"%h/.gemini/skills", ".agents/skills"},
	"copilot":  {"%h/.copilot/skills", ".agents/skills"},
	"opencode": {"%h/.config/opencode/skills", ".agents/skills"},
}

func init() {
	skillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Install this CLI's AI-agent skill into Claude, Cursor, and other agents",
		Long: `Install the n8nctl agent skill so AI coding agents know how to drive
this CLI.

The cross-agent way (recommended) is Vercel's installer, which detects every
agent you have and is always up to date:

  npx skills add jjuanrivvera/n8n-cli

This built-in command writes the bundled skill directly, without Node:

  n8nctl skills install                 # Claude Code (project ./.claude/skills)
  n8nctl skills install --global        # Claude Code (~/.claude/skills)
  n8nctl skills install --agent cursor --global`,
	}
	skillsCmd.AddCommand(newSkillsInstallCmd(), newSkillsPathCmd(), newSkillsPrintCmd())
	rootCmd.AddCommand(skillsCmd)
}

func newSkillsInstallCmd() *cobra.Command {
	var (
		global bool
		agent  string
		dir    string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Write the bundled skill into an agent's skills directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := resolveSkillTarget(agent, dir, global)
			if err != nil {
				return err
			}
			files, err := skillFiles()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if flagDryRun {
				fmt.Fprintf(out, "Would install skill to %s:\n", target)
				for _, f := range files {
					fmt.Fprintf(out, "  %s\n", filepath.Join(target, f))
				}
				return nil
			}
			for _, f := range files {
				data, rerr := fs.ReadFile(n8ncli.SkillFS, f)
				if rerr != nil {
					return rerr
				}
				dst := filepath.Join(target, filepath.FromSlash(f))
				if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { //nolint:gosec // skill dir is user-facing
					return err
				}
				if err := os.WriteFile(dst, data, 0o644); err != nil { //nolint:gosec // skill docs are world-readable
					return err
				}
			}
			fmt.Fprintf(out, "Installed n8nctl skill to %s\n", target)
			fmt.Fprintln(out, "Tip: for every agent at once, use `npx skills add jjuanrivvera/n8n-cli`.")
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Install for all projects (user-level) instead of the current project")
	cmd.Flags().StringVar(&agent, "agent", "claude", "Target agent: "+strings.Join(agentNames(), ", "))
	cmd.Flags().StringVar(&dir, "dir", "", "Explicit destination base directory (overrides --agent/--global)")
	_ = cmd.RegisterFlagCompletionFunc("agent", fixedCompletions(agentNames()))
	return cmd
}

func newSkillsPathCmd() *cobra.Command {
	var (
		global bool
		agent  string
	)
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print where the skill would be installed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := resolveSkillTarget(agent, "", global)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), target)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "User-level path")
	cmd.Flags().StringVar(&agent, "agent", "claude", "Target agent: "+strings.Join(agentNames(), ", "))
	return cmd
}

func newSkillsPrintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Print the bundled SKILL.md to stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data, err := fs.ReadFile(n8ncli.SkillFS, "SKILL.md")
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

// resolveSkillTarget computes the destination directory for the skill.
func resolveSkillTarget(agent, dir string, global bool) (string, error) {
	if dir != "" {
		return filepath.Join(dir, n8ncli.SkillName), nil
	}
	paths, ok := agentSkillDirs[strings.ToLower(agent)]
	if !ok {
		return "", fmt.Errorf("unknown agent %q (known: %s)", agent, strings.Join(agentNames(), ", "))
	}
	base := paths[1] // project
	if global {
		base = paths[0]
	}
	if strings.HasPrefix(base, "%h") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(base, "%h/")))
	}
	return filepath.Join(base, n8ncli.SkillName), nil
}

// skillFiles lists the embedded skill files (relative slash paths), files only.
func skillFiles() ([]string, error) {
	var files []string
	err := fs.WalkDir(n8ncli.SkillFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func agentNames() []string {
	names := make([]string, 0, len(agentSkillDirs))
	for n := range agentSkillDirs {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
