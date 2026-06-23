package commands

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/wffile"
	"github.com/jjuanrivvera/n8n-cli/internal/wflint"
)

// workflowAutofixCmd auto-repairs the mechanical mistakes the linter detects:
// node-type typos (corrected against the node catalog), expression strings
// missing the leading '=', and webhook/form-trigger nodes missing a webhookId.
func workflowAutofixCmd() *cobra.Command {
	var files []string
	var dir string
	var write bool

	cmd := &cobra.Command{
		Use:   "autofix [-f <file>... | --dir <dir>]",
		Short: "Auto-repair common workflow mistakes in files",
		Long: "Apply mechanical fixes to workflow files: correct typo'd node types (against\n" +
			"the embedded node catalog), add the leading '=' to expression strings that are\n" +
			"missing it, and generate a webhookId for webhook/form-trigger nodes that lack one.\n\n" +
			"By default it reports what it would change; pass --write to apply the fixes.",
		Args: cobra.NoArgs,
		Example: "  n8nctl workflows autofix --dir ./workflows\n" +
			"  n8nctl workflows autofix -f wf.json --write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			targets, err := autofixTargets(dir, files)
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				return fmt.Errorf("provide --dir or -f/--file")
			}
			out := cmd.OutOrStdout()
			totalFixed := 0
			for _, path := range targets {
				raw, rerr := os.ReadFile(path) //nolint:gosec // user-supplied path
				if rerr != nil {
					return rerr
				}
				format := wffile.FormatFromPath(path)
				wf, derr := wffile.Decode(raw, format)
				if derr != nil {
					return fmt.Errorf("%s: %w", path, derr)
				}
				fixes, ferr := autofixWorkflow(wf)
				if ferr != nil {
					return fmt.Errorf("%s: %w", path, ferr)
				}
				if len(fixes) == 0 {
					continue
				}
				totalFixed += len(fixes)
				for _, f := range fixes {
					fmt.Fprintf(out, "%s · %s\n", path, f)
				}
				if write {
					encoded, eerr := wffile.Encode(wf, format)
					if eerr != nil {
						return eerr
					}
					// #nosec G703 G304 -- writes back the same user-supplied file path, not untrusted input
					if werr := os.WriteFile(path, encoded, 0o600); werr != nil {
						return werr
					}
				}
			}
			verb := "would fix"
			if write {
				verb = "fixed"
			}
			if totalFixed == 0 {
				fmt.Fprintln(out, "nothing to fix")
			} else if !write {
				fmt.Fprintf(out, "\n%s %d issue(s) — re-run with --write to apply\n", verb, totalFixed)
			} else {
				fmt.Fprintf(out, "\n%s %d issue(s)\n", verb, totalFixed)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "workflow files to fix")
	cmd.Flags().StringVar(&dir, "dir", "", "fix all workflow files in a directory")
	cmd.Flags().BoolVar(&write, "write", false, "write the fixes back (default: report only)")
	return cmd
}

func autofixTargets(dir string, files []string) ([]string, error) {
	if dir != "" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		var out []string
		for _, e := range entries {
			if !e.IsDir() && isWorkflowFile(e.Name()) {
				out = append(out, dir+string(os.PathSeparator)+e.Name())
			}
		}
		return out, nil
	}
	return files, nil
}

// autofixWorkflow mutates wf in place and returns a description of each fix.
func autofixWorkflow(wf *api.Workflow) ([]string, error) {
	var nodes []map[string]any
	if err := json.Unmarshal(wf.Nodes, &nodes); err != nil {
		return nil, nil // not auto-fixable; leave for lint to report
	}
	var fixes []string
	for _, n := range nodes {
		name, _ := n["name"].(string)
		nodeType, _ := n["type"].(string)

		// 1. node-type typo
		if corrected, ok := wflint.NodeTypeCorrection(nodeType); ok {
			n["type"] = corrected
			fixes = append(fixes, fmt.Sprintf("%s: node type %q -> %q", name, nodeType, corrected))
			nodeType = corrected
		}
		// 2. missing webhookId
		if wflint.IsWebhookNode(nodeType) {
			if id, _ := n["webhookId"].(string); id == "" {
				wid, err := newWebhookID()
				if err != nil {
					return nil, err
				}
				n["webhookId"] = wid
				fixes = append(fixes, fmt.Sprintf("%s: generated webhookId", name))
			}
		}
		// 3. expression strings missing the leading '='
		if params, ok := n["parameters"].(map[string]any); ok {
			c := fixExpressions(params)
			if c > 0 {
				fixes = append(fixes, fmt.Sprintf("%s: prefixed %d expression(s) with '='", name, c))
			}
		}
	}
	if len(fixes) == 0 {
		return nil, nil
	}
	patched, err := json.Marshal(nodes)
	if err != nil {
		return nil, err
	}
	wf.Nodes = patched
	return fixes, nil
}

// fixExpressions rewrites every nested string that is an unprefixed n8n expression
// to start with '=', returning the number of strings changed.
func fixExpressions(v any) int {
	count := 0
	switch t := v.(type) {
	case map[string]any:
		for k, vv := range t {
			if s, ok := vv.(string); ok {
				if wflint.NeedsExpressionPrefix(s) {
					t[k] = "=" + s
					count++
				}
				continue
			}
			count += fixExpressions(vv)
		}
	case []any:
		for i, vv := range t {
			if s, ok := vv.(string); ok {
				if wflint.NeedsExpressionPrefix(s) {
					t[i] = "=" + s
					count++
				}
				continue
			}
			count += fixExpressions(vv)
		}
	}
	return count
}

// newWebhookID returns a random UUIDv4 string, the format n8n uses for webhookId.
func newWebhookID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
