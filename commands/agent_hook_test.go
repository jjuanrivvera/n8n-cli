package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHookScript_BashExecution exercises the generated hook script with real
// bash to verify the adversarial cases from the agent-guard audit. Gated on
// runtime.GOOS != "windows" and bash being available, so it is safe to include
// in the regular test suite.
func TestHookScript_BashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not found in PATH; skipping hook execution tests")
	}

	// Generate the hook from the real command tree so the blocked_cmds /
	// blocked_tools arrays are fully populated.
	_, _, irreversible := classifyAPICommands(RootCmd())

	hookContent := claudeHookScript(irreversible)
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "n8nctl-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { //nolint:gosec // executable hook
		t.Fatalf("write hook: %v", err)
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  toolName,
			"tool_input": map[string]any{},
		})
		return string(b)
	}

	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		// The hook always exits 0; the decision is in the JSON output.
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}
	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name        string
		payload     string
		wantDenied  bool
		description string
	}{
		{
			name:        "bash_workflows_delete_denied",
			payload:     bashPayload("n8nctl workflows delete 42"),
			wantDenied:  true,
			description: "direct n8nctl workflows delete must be denied",
		},
		{
			name:        "bash_executions_prune_denied",
			payload:     bashPayload("n8nctl executions prune --older-than 30d --yes"),
			wantDenied:  true,
			description: "n8nctl executions prune must be denied",
		},
		{
			name:        "bash_compound_delete_rows_denied",
			payload:     bashPayload("n8nctl data-tables delete-rows 7 --data '{}'"),
			wantDenied:  true,
			description: "compound delete-rows path must be denied",
		},
		{
			name:        "bash_alias_wf_delete_denied",
			payload:     bashPayload("n8nctl wf delete 42"),
			wantDenied:  true,
			description: "cobra alias spelling wf delete must be denied",
		},
		{
			name:        "bash_alias_exec_prune_denied",
			payload:     bashPayload("n8nctl exec prune --yes"),
			wantDenied:  true,
			description: "cobra alias spelling exec prune must be denied",
		},
		{
			name:        "bash_obfuscated_delete_denied",
			payload:     bashPayload(`n8nctl workflows de""lete 42`),
			wantDenied:  true,
			description: "quote-obfuscated delete must be denied",
		},
		{
			name:        "bash_after_semicolon_denied",
			payload:     bashPayload("echo hi; n8nctl workflows delete 42"),
			wantDenied:  true,
			description: "blocked command after ; must be denied",
		},
		{
			name:        "bash_env_prefix_denied",
			payload:     bashPayload("env X=1 n8nctl workflows delete 42"),
			wantDenied:  true,
			description: "env-prefixed invocation must be denied",
		},
		{
			name:        "bash_relative_path_binary_denied",
			payload:     bashPayload("./bin/n8nctl workflows delete 42"),
			wantDenied:  true,
			description: "path-invoked binary ./bin/n8nctl must be denied",
		},
		{
			name:        "bash_absolute_path_binary_denied",
			payload:     bashPayload("/usr/local/bin/n8nctl workflows delete 42"),
			wantDenied:  true,
			description: "path-invoked binary /usr/local/bin/n8nctl must be denied",
		},
		{
			name:        "bash_api_delete_denied",
			payload:     bashPayload("n8nctl api DELETE /workflows/42"),
			wantDenied:  true,
			description: "raw-api DELETE must be denied",
		},
		{
			name:        "bash_api_delete_lowercase_denied",
			payload:     bashPayload("n8nctl api delete /workflows/42"),
			wantDenied:  true,
			description: "raw-api lowercase delete must be denied (method is upcased by the CLI)",
		},
		{
			name:        "bash_api_post_denied",
			payload:     bashPayload(`n8nctl api POST /workflows -d '{}'`),
			wantDenied:  true,
			description: "raw-api POST must be denied",
		},
		{
			name:        "bash_api_put_denied",
			payload:     bashPayload(`n8nctl api PUT /workflows/42 -d '{}'`),
			wantDenied:  true,
			description: "raw-api PUT must be denied",
		},
		{
			name:        "bash_path_invoked_api_delete_denied",
			payload:     bashPayload("/usr/local/bin/n8nctl api DELETE /workflows/42"),
			wantDenied:  true,
			description: "path-invoked raw-api DELETE must be denied",
		},
		{
			name:        "bash_read_allowed",
			payload:     bashPayload("n8nctl workflows list --all"),
			wantDenied:  false,
			description: "read command must NOT be denied",
		},
		{
			name:        "bash_create_with_delete_arg_allowed",
			payload:     bashPayload("n8nctl workflows create --set name=delete-old-data"),
			wantDenied:  false,
			description: "create with a benign 'delete' in an argument must NOT be denied",
		},
		{
			name:        "bash_unrelated_file_allowed",
			payload:     bashPayload("cat workflows_delete.go"),
			wantDenied:  false,
			description: "non-n8nctl command mentioning 'delete' must NOT be denied",
		},
		{
			name:        "bash_api_get_with_delete_in_path_allowed",
			payload:     bashPayload("n8nctl api GET /workflows?filter=delete"),
			wantDenied:  false,
			description: "raw-api GET whose PATH contains 'delete' must NOT be denied",
		},
		{
			name:        "bash_other_binary_named_like_n8nctl_allowed",
			payload:     bashPayload("myn8nctl workflows delete 42"),
			wantDenied:  false,
			description: "a different binary whose name merely ends in 'n8nctl' must NOT be denied",
		},
		{
			name:        "bash_glued_separator_denied",
			payload:     bashPayload("n8nctl executions prune;true"),
			wantDenied:  true,
			description: "shell separator glued directly to the verb must still be denied",
		},
		{
			name:        "bash_glued_pipe_denied",
			payload:     bashPayload("n8nctl workflows delete|cat"),
			wantDenied:  true,
			description: "pipe glued directly to the verb must still be denied",
		},
		{
			// Conservative false positive, accepted: quote-stripping
			// de-obfuscation makes the quoted search string indistinguishable
			// from a real invocation at a command position. Denying is the
			// safe direction; locked in so a change here is a conscious one.
			name:        "bash_rg_quoted_blocked_string_denied_conservative",
			payload:     bashPayload(`rg "n8nctl workflows delete" src/`),
			wantDenied:  true,
			description: "quoted blocked string in a search arg is denied (accepted conservative FP)",
		},
		{
			name:        "mcp_workflows_delete_denied",
			payload:     mcpPayload("mcp__n8nctl__n8n_workflows_delete"),
			wantDenied:  true,
			description: "MCP n8n_workflows_delete tool must be denied",
		},
		{
			name:        "mcp_renamed_server_still_denied",
			payload:     mcpPayload("mcp__my-n8n__n8n_workflows_delete"),
			wantDenied:  true,
			description: "MCP tool basename match must cover a renamed server",
		},
		{
			name:        "mcp_workflows_list_allowed",
			payload:     mcpPayload("mcp__n8nctl__n8n_workflows_list"),
			wantDenied:  false,
			description: "MCP n8n_workflows_list tool must NOT be denied",
		},
		{
			name:        "mcp_near_miss_tool_allowed",
			payload:     mcpPayload("mcp__n8nctl__n8n_workflows_delete2"),
			wantDenied:  false,
			description: "near-miss MCP tool name (suffix added) must NOT be denied",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			denied := isDenied(output)
			if denied != tc.wantDenied {
				want, got := "allowed", "denied"
				if tc.wantDenied {
					want, got = "denied", "allowed"
				}
				t.Errorf("%s: want %s, got %s\noutput: %s", tc.description, want, got, output)
			}
		})
	}
}
