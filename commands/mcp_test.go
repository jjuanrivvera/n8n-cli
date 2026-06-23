package commands

import (
	"testing"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findCmd walks the command tree to the command at the given path.
func findCmd(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	c := RootCmd()
	for _, name := range path {
		var next *cobra.Command
		for _, sub := range c.Commands() {
			if sub.Name() == name {
				next = sub
				break
			}
		}
		require.NotNil(t, next, "command %v not found (missing %q)", path, name)
		c = next
	}
	return c
}

// TestMCPExcludesSetupCommands locks the MCP tool surface to n8n operations:
// local setup/meta commands must not be exposed as tools, while resource
// operations (including destructive ones) must remain reachable.
func TestMCPExcludesSetupCommands(t *testing.T) {
	exclude := ophis.ExcludeCmdsContaining(mcpExcludedCommands...)

	excluded := [][]string{
		{"agent", "guard"}, {"auth", "login"}, {"config", "use"},
		{"alias", "set"}, {"init"}, {"doctor"},
	}
	for _, p := range excluded {
		c := findCmd(t, p...)
		assert.False(t, exclude(c), "%v must be excluded from the MCP tool surface", p)
	}

	// exclude() returns true for commands that PASS the filter (stay exposed).
	included := [][]string{
		{"workflows", "list"}, {"workflows", "delete"}, {"credentials", "create"},
		{"executions", "retry"},
	}
	for _, p := range included {
		c := findCmd(t, p...)
		assert.True(t, exclude(c), "%v must stay in the MCP tool surface", p)
	}
}

// TestMCPHints verifies the read-only/write/destructive annotations that ophis
// surfaces on every tool and that agent guard uses to classify operations.
func TestMCPHints(t *testing.T) {
	cases := []struct {
		path     []string
		readOnly bool
		destruct bool
	}{
		{[]string{"workflows", "list"}, true, false},
		{[]string{"workflows", "get"}, true, false},
		{[]string{"workflows", "create"}, false, false}, // write: openWorld only
		{[]string{"workflows", "delete"}, false, true},
		{[]string{"workflows", "activate"}, false, true}, // custom verb -> destructive default
		{[]string{"workflows", "search"}, true, false},   // read-only custom verb
		{[]string{"credentials", "schema"}, true, false},
	}
	for _, c := range cases {
		cmd := findCmd(t, c.path...)
		assert.Equal(t, c.readOnly, cmd.Annotations[ophis.AnnotationReadOnly] == "true", "%v readOnly", c.path)
		assert.Equal(t, c.destruct, cmd.Annotations[ophis.AnnotationDestructive] == "true", "%v destructive", c.path)
		assert.Equal(t, "true", cmd.Annotations[ophis.AnnotationOpenWorld], "%v must be an API op", c.path)
	}
}
