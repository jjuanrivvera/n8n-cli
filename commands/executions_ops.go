package commands

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// executionPruneCmd bulk-deletes executions by age and/or status, to reclaim n8n
// database space on busy instances (the UI deletes one page at a time).
func executionPruneCmd() *cobra.Command {
	var olderThan, status, workflow string
	var yes bool
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Bulk-delete executions by age and/or status",
		Long: "Delete execution records older than a cutoff and/or matching a status, to\n" +
			"reclaim database space. Always previews the count first; pass --yes to skip the\n" +
			"confirmation, or --dry-run to only count.",
		Args: cobra.NoArgs,
		Example: "  n8nctl executions prune --older-than 30d\n" +
			"  n8nctl executions prune --older-than 7d --status error --yes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if olderThan == "" && status == "" {
				return fmt.Errorf("provide --older-than and/or --status")
			}
			cutoff, err := ageCutoff(olderThan)
			if err != nil {
				return err
			}
			client, err := getReadClient(cmd)
			if err != nil {
				return err
			}
			params := api.ListParams{Extra: url.Values{}}
			if status != "" {
				params.Extra.Set("status", status)
			}
			if workflow != "" {
				params.Extra.Set("workflowId", workflow)
			}
			all, err := client.Executions().ListAll(cmd.Context(), params, 0)
			if err != nil {
				return err
			}
			var victims []api.Execution
			for i := range all {
				if olderThan != "" {
					t, perr := time.Parse(time.RFC3339, all[i].StartedAt)
					if perr != nil || !t.Before(cutoff) {
						continue
					}
				}
				victims = append(victims, all[i])
			}
			out := cmd.OutOrStdout()
			if len(victims) == 0 {
				fmt.Fprintln(out, "nothing to prune")
				return nil
			}
			if flagDryRun {
				fmt.Fprintf(out, "would delete %d execution(s)\n", len(victims))
				return nil
			}
			if !yes && !confirm(cmd, fmt.Sprintf("delete %d execution(s)?", len(victims))) {
				fmt.Fprintln(out, "aborted")
				return nil
			}
			deleted, failed := 0, 0
			for i := range victims {
				if derr := client.Executions().Delete(cmd.Context(), victims[i].ID.String()); derr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "failed to delete %s: %v\n", victims[i].ID, derr)
					failed++
					continue
				}
				deleted++
			}
			fmt.Fprintf(out, "pruned %d execution(s)", deleted)
			if failed > 0 {
				fmt.Fprintf(out, ", %d failed", failed)
			}
			fmt.Fprintln(out)
			if failed > 0 {
				return fmt.Errorf("%d execution(s) could not be deleted", failed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&olderThan, "older-than", "", "delete executions older than this (e.g. 30d, 720h, 90m)")
	cmd.Flags().StringVar(&status, "status", "", "only delete this status (error, success, ...)")
	cmd.Flags().StringVar(&workflow, "workflow", "", "only delete executions of this workflow id")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}

// executionWatchCmd live-tails new executions, highlighting failures — a
// `kubectl get -w` for runs.
func executionWatchCmd() *cobra.Command {
	var status, workflow string
	var interval time.Duration
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Live-tail new executions, highlighting failures",
		Long: "Poll the executions endpoint and print each new run as it appears, coloring\n" +
			"failures. Runs until interrupted (Ctrl-C).",
		Args: cobra.NoArgs,
		Example: "  n8nctl executions watch\n" +
			"  n8nctl executions watch --status error --interval 10s",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := getReadClient(cmd)
			if err != nil {
				return err
			}
			params := api.ListParams{Limit: 50, Extra: url.Values{}}
			if status != "" {
				params.Extra.Set("status", status)
			}
			if workflow != "" {
				params.Extra.Set("workflowId", workflow)
			}
			out := cmd.OutOrStdout()
			color := !flagNoColor && os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd()))
			seen := map[string]bool{}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			fmt.Fprintf(cmd.ErrOrStderr(), "watching executions (every %s) — Ctrl-C to stop\n", interval)

			first := true
			for {
				items, _, lerr := client.Executions().List(cmd.Context(), params)
				if lerr != nil {
					if cmd.Context().Err() != nil {
						return nil
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "poll error: %v\n", lerr)
				}
				// The API returns newest first; collect unseen, then print oldest-first.
				var fresh []api.Execution
				for i := range items {
					id := items[i].ID.String()
					if !seen[id] {
						seen[id] = true
						fresh = append(fresh, items[i])
					}
				}
				if !first {
					for i := len(fresh) - 1; i >= 0; i-- {
						fmt.Fprintln(out, formatExecLine(fresh[i], color))
					}
				}
				first = false
				select {
				case <-cmd.Context().Done():
					return nil
				case <-ticker.C:
				}
			}
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "only watch this status")
	cmd.Flags().StringVar(&workflow, "workflow", "", "only watch executions of this workflow id")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "poll interval")
	return cmd
}

// ageCutoff turns "30d"/"720h"/"90m" into the time before which records are old.
func ageCutoff(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if days, ok := strings.CutSuffix(s, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil || n < 0 {
			return time.Time{}, fmt.Errorf("invalid --older-than %q (e.g. 30d)", s)
		}
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --older-than %q (e.g. 30d, 720h, 90m)", s)
	}
	return time.Now().Add(-d), nil
}

func formatExecLine(e api.Execution, color bool) string {
	st := e.Status
	if color {
		switch st {
		case "error", "crashed", "canceled":
			st = "\x1b[31m" + st + "\x1b[0m" // red
		case "success":
			st = "\x1b[32m" + st + "\x1b[0m" // green
		case "running", "waiting":
			st = "\x1b[33m" + st + "\x1b[0m" // yellow
		}
	}
	return fmt.Sprintf("%s  wf=%s  %s  %s", e.StartedAt, e.WorkflowID, st, e.ID)
}
