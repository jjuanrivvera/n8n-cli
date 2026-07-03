package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	pkg := &cobra.Command{
		Use:     "packages",
		Aliases: []string{"package", "pkg"},
		Short:   "Export and import workflows as .n8np packages (beta)",
		Long: "Bundle workflows into a portable .n8np archive and import them elsewhere.\n" +
			"This is a beta n8n feature, disabled by default; the API returns 404 unless\n" +
			"the instance sets N8N_PUBLIC_API_PACKAGES_ENABLED=true.",
	}
	// export only reads the instance (the archive lands in a local file);
	// import CREATES workflows/credentials on the instance — a remote write that
	// must carry openWorldHint so `n8nctl agent guard` gates it (an unannotated
	// top-level command is treated as local/utility and never gated).
	pkg.AddCommand(readOnlyHints(packageExportCmd()), writeHints(packageImportCmd()))
	rootCmd.AddCommand(pkg)
}

func packageExportCmd() *cobra.Command {
	var ids []string
	var out string
	cmd := &cobra.Command{
		Use:   "export --workflow <id> [--workflow <id> ...] --out <file.n8np>",
		Short: "Export workflows as a .n8np package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(ids) == 0 {
				return fmt.Errorf("at least one --workflow id is required")
			}
			if out == "" {
				return fmt.Errorf("--out <file> is required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			archive, err := client.ExportPackage(cmd.Context(), ids)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if err := os.WriteFile(out, archive, 0o600); err != nil {
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "exported %d workflow(s) to %s (%d bytes)\n", len(ids), out, len(archive))
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&ids, "workflow", nil, "workflow id to export (repeatable)")
	cmd.Flags().StringVar(&out, "out", "", "output .n8np file (required)")
	return cmd
}

func packageImportCmd() *cobra.Command {
	var file string
	var opts api.ImportOptions
	cmd := &cobra.Command{
		Use:   "import --file <file.n8np> --conflict-policy <policy>",
		Short: "Import a .n8np package into a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if file == "" {
				return fmt.Errorf("--file <file.n8np> is required")
			}
			if opts.ConflictPolicy == "" {
				return fmt.Errorf("--conflict-policy is required (e.g. fail, new-version)")
			}
			archive, err := os.ReadFile(file) //nolint:gosec // user-supplied package path
			if err != nil {
				return err
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			res, err := client.ImportPackage(cmd.Context(), archive, opts)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, res)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to the .n8np package (required)")
	cmd.Flags().StringVar(&opts.ConflictPolicy, "conflict-policy", "", "workflow conflict policy (required), e.g. fail|new-version")
	cmd.Flags().StringVar(&opts.ProjectID, "project", "", "destination project id (default: personal project)")
	cmd.Flags().StringVar(&opts.FolderID, "folder", "", "destination folder id")
	cmd.Flags().StringVar(&opts.WorkflowIDPolicy, "workflow-id-policy", "", "workflow id policy, e.g. new")
	cmd.Flags().StringVar(&opts.CredentialMatchingMode, "credential-matching-mode", "", "credential matching mode (id-only)")
	cmd.Flags().StringVar(&opts.CredentialMissingMode, "credential-missing-mode", "", "credential missing mode")
	return cmd
}
