package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
)

func TestToolsListIncludesLaunchCodex(t *testing.T) {
	server := Server{}
	responsePayload, err := server.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var got struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(responsePayload, &got); err != nil {
		t.Fatalf("response is invalid json: %v", err)
	}

	names := map[string]bool{}
	for _, tool := range got.Result.Tools {
		names[tool.Name] = true
	}
	if !names["vmware_mcp_launch_codex"] {
		t.Fatalf("launch codex tool missing: %#v", names)
	}
	if !names["vmware_mcp_run_command"] {
		t.Fatalf("run command tool missing: %#v", names)
	}
}

func TestLaunchAppQueuesPendingRequest(t *testing.T) {
	root := t.TempDir()
	server := Server{
		Queue: queue.New(root),
		Now:   func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) },
	}

	payload := []byte(`{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "tools/call",
		"params": {
			"name": "vmware_mcp_launch_app",
			"arguments": {"app_name": "Codex"}
		}
	}`)
	responsePayload, err := server.Handle(context.Background(), payload)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var got struct {
		Result struct {
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(responsePayload, &got); err != nil {
		t.Fatalf("response is invalid json: %v", err)
	}
	if got.Result.IsError {
		t.Fatal("result is error")
	}

	matches, err := filepath.Glob(filepath.Join(root, "requests", "*.pending.json"))
	if err != nil {
		t.Fatalf("glob returned error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("pending requests = %d, want 1", len(matches))
	}
	payload, err = os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read pending request: %v", err)
	}
	if !json.Valid(payload) {
		t.Fatalf("pending request is not json: %s", payload)
	}
}
