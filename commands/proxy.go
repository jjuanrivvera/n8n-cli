package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/wflint"
)

func init() {
	rootCmd.AddCommand(proxyCmd())
}

func proxyCmd() *cobra.Command {
	var listen string
	var disable []string
	var blockDestructive bool

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Run a local n8n API proxy that lint-gates workflow writes",
		Long: "Stand a reverse proxy in front of the active instance that lints every\n" +
			"workflow create/update and rejects failures with HTTP 422 before they reach\n" +
			"n8n. This makes linting structural rather than a convention: anything that\n" +
			"pushes a workflow through the proxy — a human, a script, or an AI agent —\n" +
			"is held to the same rules, so a definition with errors can never land.\n\n" +
			"Point your n8n client at the proxy as if it were the instance host:\n" +
			"  n8nctl proxy &                       # listens on 127.0.0.1:8099\n" +
			"  export N8N_API_URL=http://127.0.0.1:8099   # for any n8n client\n" +
			"  n8nctl --base-url http://127.0.0.1:8099 workflows create --file wf.json\n\n" +
			"Reads pass straight through. The proxy injects the active profile's API key\n" +
			"(from your keyring), so the client never needs it. Bind to localhost only\n" +
			"unless you understand that the proxy is an authenticated gateway to n8n.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			baseURL, apiKey, _, err := resolveTarget()
			if err != nil {
				return err
			}
			u, err := url.Parse(baseURL)
			if err != nil {
				return fmt.Errorf("invalid base URL %q: %w", baseURL, err)
			}
			// The proxy stands in for the n8n host; the /api/v1 path comes from the
			// client's request, so target only the scheme+host.
			target := &url.URL{Scheme: u.Scheme, Host: u.Host}

			disabled := map[string]bool{}
			for _, d := range disable {
				disabled[d] = true
			}

			handler := newLintProxy(target, apiKey, disabled, blockDestructive, cmd.ErrOrStderr())
			srv := &http.Server{Addr: listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second}

			// Shut down cleanly on Ctrl-C (cmd.Context() is signal-cancelled).
			go func() {
				<-cmd.Context().Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()

			fmt.Fprintf(cmd.ErrOrStderr(),
				"n8nctl proxy listening on http://%s → %s\n  workflow writes are lint-gated; reads pass through. Ctrl-C to stop.\n",
				listen, target)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&listen, "listen", "127.0.0.1:8099", "address to listen on")
	cmd.Flags().StringSliceVar(&disable, "disable-rule", nil, "lint rules to disable (comma-separated)")
	cmd.Flags().BoolVar(&blockDestructive, "block-destructive", false, "also reject workflow DELETE requests")
	return cmd
}

// newLintProxy returns a reverse proxy to target that lints workflow create/update
// bodies and rejects errors with 422, optionally blocking workflow DELETEs. It
// injects apiKey as X-N8N-API-KEY so clients forward without the secret.
func newLintProxy(target *url.URL, apiKey string, disabled map[string]bool, blockDestructive bool, logw io.Writer) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(target)
	director := rp.Director
	rp.Director = func(req *http.Request) {
		director(req)
		req.Host = target.Host
		req.Header.Set("X-N8N-API-KEY", apiKey)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case isWorkflowWrite(r):
			body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20)) // 16 MiB cap
			_ = r.Body.Close()
			if err != nil {
				writeProxyError(w, http.StatusBadRequest, "n8nctl proxy: could not read request body", nil)
				return
			}
			if findings := lintWriteBody(body, disabled); len(findings) > 0 {
				fmt.Fprintf(logw, "✗ %s %s → 422 (%d lint error(s))\n", r.Method, r.URL.Path, len(findings))
				writeProxyError(w, http.StatusUnprocessableEntity,
					"n8nctl proxy: workflow rejected by lint", findings)
				return
			}
			fmt.Fprintf(logw, "✓ %s %s → forwarded (lint clean)\n", r.Method, r.URL.Path)
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			rp.ServeHTTP(w, r)
		case blockDestructive && isWorkflowDelete(r):
			fmt.Fprintf(logw, "✗ %s %s → 403 (--block-destructive)\n", r.Method, r.URL.Path)
			writeProxyError(w, http.StatusForbidden,
				"n8nctl proxy: workflow DELETE blocked (--block-destructive)", nil)
		default:
			rp.ServeHTTP(w, r)
		}
	})
}

// isWorkflowWrite reports whether r creates (POST .../workflows) or updates
// (PUT/PATCH .../workflows/{id}) a full workflow — the bodies worth linting. It
// deliberately excludes sub-resources like .../workflows/{id}/tags|activate.
func isWorkflowWrite(r *http.Request) bool {
	rest, ok := afterWorkflows(r.URL.Path)
	if !ok {
		return false
	}
	switch r.Method {
	case http.MethodPost:
		return len(rest) == 0 // .../workflows
	case http.MethodPut, http.MethodPatch:
		return len(rest) == 1 // .../workflows/{id}
	default:
		return false
	}
}

// isWorkflowDelete reports whether r deletes a workflow (DELETE .../workflows/{id}).
func isWorkflowDelete(r *http.Request) bool {
	rest, ok := afterWorkflows(r.URL.Path)
	return ok && r.Method == http.MethodDelete && len(rest) == 1
}

// afterWorkflows returns the path segments after the "workflows" segment.
func afterWorkflows(path string) ([]string, bool) {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i, s := range segs {
		if s == "workflows" {
			return segs[i+1:], true
		}
	}
	return nil, false
}

// lintWriteBody runs the lint engine over a workflow write body and returns the
// error-severity findings (warnings do not block, matching `workflows lint`).
func lintWriteBody(body []byte, disabled map[string]bool) []wflint.Finding {
	var wf api.Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		// Not a parseable workflow — let n8n return its own validation error.
		return nil
	}
	var errs []wflint.Finding
	for _, f := range wflint.Lint(&wf, disabled) {
		if f.Severity == wflint.Error {
			errs = append(errs, f)
		}
	}
	return errs
}

// writeProxyError writes a JSON error body that mirrors n8n's {message, ...} shape.
func writeProxyError(w http.ResponseWriter, code int, message string, findings []wflint.Finding) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	payload := map[string]any{"message": message}
	if findings != nil {
		payload["lint"] = findings
	}
	_ = json.NewEncoder(w).Encode(payload)
}
