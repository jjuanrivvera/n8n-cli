package api

import (
	"context"
	"encoding/json"
)

// Execution models an n8n workflow execution. Note that id and workflowId are
// numeric on the wire (sometimes quoted), which the flexible ID type absorbs.
//
// See https://docs.n8n.io/api/api-reference/#tag/Execution
type Execution struct {
	ID             ID              `json:"id,omitempty"`
	WorkflowID     ID              `json:"workflowId,omitempty"`
	Finished       Bool            `json:"finished,omitempty"`
	Mode           string          `json:"mode,omitempty"`
	Status         string          `json:"status,omitempty"`
	RetryOf        ID              `json:"retryOf,omitempty"`
	RetrySuccessID ID              `json:"retrySuccessId,omitempty"`
	StartedAt      string          `json:"startedAt,omitempty"`
	StoppedAt      string          `json:"stoppedAt,omitempty"`
	WaitTill       string          `json:"waitTill,omitempty"`
	CustomData     json.RawMessage `json:"customData,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
}

// Executions returns a typed handle to the /executions resource. Executions are
// read-only with custom actions (retry/stop) — no create or update.
func (c *Client) Executions() *Resource[Execution] { return NewResource[Execution](c, "executions") }

// RetryExecution re-runs a failed execution (POST /executions/{id}/retry).
// loadWorkflow re-loads the current workflow definition instead of the one used
// at execution time.
func (c *Client) RetryExecution(ctx context.Context, id string, loadWorkflow bool) (*Execution, error) {
	var out Execution
	body := map[string]bool{"loadWorkflow": loadWorkflow}
	err := c.Executions().Action(ctx, id, "retry", body, &out)
	return &out, err
}

// StopExecution stops a running execution (POST /executions/{id}/stop).
func (c *Client) StopExecution(ctx context.Context, id string) (*Execution, error) {
	var out Execution
	err := c.Executions().Action(ctx, id, "stop", nil, &out)
	return &out, err
}
