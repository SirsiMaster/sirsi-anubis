package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	WakeCLISpawn         = "cli-spawn"
	WakeAPICall          = "api-call"
	WakeMCPNotification  = "mcp-notification"
	mcpNotificationFile  = "mcp-notifications.jsonl"
	defaultWakeHTTPDelay = 15 * time.Second
)

// WakePayload is the normalized message sent through non-CLI wake adapters.
type WakePayload struct {
	AgentID  string `json:"agent_id"`
	Type     string `json:"type"`
	DocID    string `json:"doc_id"`
	Topic    string `json:"topic,omitempty"`
	Goal     string `json:"goal,omitempty"`
	RepoRoot string `json:"repo_root,omitempty"`
	Prompt   string `json:"prompt"`
}

type wakeHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (e *Executor) wake(ctx context.Context, item *WorkItem, cfg *AgentConfig, prompt string) error {
	switch cfg.WakeMechanism() {
	case WakeAPICall:
		return e.wakeAPI(ctx, item, cfg, prompt)
	case WakeMCPNotification:
		return e.wakeMCP(ctx, item, cfg, prompt)
	case WakeCLISpawn, "":
		return e.wakeCLI(ctx, item, cfg, prompt)
	default:
		return fmt.Errorf("unsupported wake mechanism %q", cfg.WakeMechanism())
	}
}

func (e *Executor) wakePayload(item *WorkItem, cfg *AgentConfig, prompt string) WakePayload {
	return WakePayload{
		AgentID:  cfg.ID,
		Type:     cfg.Type,
		DocID:    item.DocID,
		Topic:    item.Topic,
		Goal:     item.Goal,
		RepoRoot: cfg.Cwd,
		Prompt:   prompt,
	}
}

func (e *Executor) wakeAPI(ctx context.Context, item *WorkItem, cfg *AgentConfig, prompt string) error {
	body, err := json.Marshal(e.wakePayload(item, cfg, prompt))
	if err != nil {
		return fmt.Errorf("marshal wake payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Wake.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create wake request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token, ok := resolveWakeAuth(cfg.Wake.Auth); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := e.wakeHTTPClient
	if client == nil {
		client = &http.Client{Timeout: defaultWakeHTTPDelay}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send wake request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wake request returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (e *Executor) wakeMCP(_ context.Context, item *WorkItem, cfg *AgentConfig, prompt string) error {
	payload := struct {
		MCPServer string      `json:"mcp_server"`
		CreatedAt string      `json:"created_at"`
		Payload   WakePayload `json:"payload"`
	}{
		MCPServer: cfg.Wake.MCPServer,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Payload:   e.wakePayload(item, cfg, prompt),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal MCP notification: %w", err)
	}
	path := filepath.Join(e.router.root, mcpNotificationFile)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open MCP notification outbox: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write MCP notification outbox: %w", err)
	}
	return nil
}

func resolveWakeAuth(spec string) (string, bool) {
	switch {
	case spec == "":
		return "", false
	case strings.HasPrefix(spec, "env:"):
		token := os.Getenv(strings.TrimPrefix(spec, "env:"))
		return token, token != ""
	case strings.HasPrefix(spec, "bearer:"):
		return strings.TrimPrefix(spec, "bearer:"), true
	default:
		return spec, true
	}
}
