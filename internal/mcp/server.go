package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
	"github.com/MarlonJD/vmware-mcp-server/internal/vmrun"
)

type Server struct {
	Queue  queue.Store
	VMRun  vmrun.Runner
	Now    func() time.Time
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *responseError `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func (server Server) Handle(ctx context.Context, payload []byte) ([]byte, error) {
	var req request
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, err
	}

	var result any
	var err error
	switch req.Method {
	case "initialize":
		result = map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "vmware-mcp-server", "version": "0.1.0"},
		}
	case "tools/list":
		result = map[string]any{"tools": tools()}
	case "tools/call":
		result, err = server.callTool(ctx, req.Params)
	default:
		return marshal(response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &responseError{Code: -32601, Message: "method not found"},
		})
	}
	if err != nil {
		return marshal(response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"isError": true,
				"content": []map[string]string{{
					"type": "text",
					"text": err.Error(),
				}},
			},
		})
	}
	return marshal(response{JSONRPC: "2.0", ID: req.ID, Result: result})
}

func marshal(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func tools() []tool {
	return []tool{
		{
			Name:        "vmware_mcp_launch_app",
			Description: "Enqueue a Windows app launch request for the guest agent.",
			InputSchema: objectSchema(map[string]any{"app_name": stringSchema()}, []string{"app_name"}),
		},
		{
			Name:        "vmware_mcp_launch_codex",
			Description: "Enqueue a Codex launch request for the Windows guest agent.",
			InputSchema: objectSchema(map[string]any{}, nil),
		},
		{
			Name:        "vmware_mcp_run_command",
			Description: "Enqueue a command request for the Windows guest agent.",
			InputSchema: objectSchema(
				map[string]any{
					"command": stringSchema(),
					"args":    map[string]any{"type": "array", "items": stringSchema()},
					"cwd":     stringSchema(),
					"admin":   map[string]any{"type": "boolean"},
				},
				[]string{"command"},
			),
		},
		{
			Name:        "vmware_mcp_get_guest_ip",
			Description: "Get the VMware guest IP address through vmrun.",
			InputSchema: objectSchema(map[string]any{"wait": map[string]any{"type": "boolean"}}, nil),
		},
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

type toolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (server Server) callTool(ctx context.Context, payload json.RawMessage) (any, error) {
	var call toolCall
	if err := json.Unmarshal(payload, &call); err != nil {
		return nil, err
	}

	switch call.Name {
	case "vmware_mcp_launch_app":
		var args struct {
			AppName string   `json:"app_name"`
			Args    []string `json:"args"`
		}
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, err
		}
		if args.AppName == "" {
			return nil, errors.New("app_name is required")
		}
		return server.enqueue(protocol.NewLaunchApp(args.AppName, args.Args))
	case "vmware_mcp_launch_codex":
		return server.enqueue(protocol.NewLaunchApp("Codex", nil))
	case "vmware_mcp_run_command":
		var args struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			CWD     string   `json:"cwd"`
			Admin   bool     `json:"admin"`
		}
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		if server.Now != nil {
			now = server.Now().UTC()
		}
		return server.enqueue(protocol.Request{
			SchemaVersion: 1,
			ID:            protocol.NewID("cmd"),
			Kind:          protocol.RequestRunCommand,
			CreatedAt:     now,
			ExpiresAt:     now.Add(2 * time.Minute),
			Command:        &protocol.RunCommand{Command: args.Command, Args: args.Args, CWD: args.CWD, Admin: args.Admin},
		})
	case "vmware_mcp_get_guest_ip":
		var args struct {
			Wait bool `json:"wait"`
		}
		_ = json.Unmarshal(call.Arguments, &args)
		output, err := server.VMRun.GetGuestIP(ctx, args.Wait)
		if err != nil {
			return nil, err
		}
		return textResult(string(output)), nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", call.Name)
	}
}

func (server Server) enqueue(request protocol.Request) (any, error) {
	now := time.Now().UTC()
	if server.Now != nil {
		now = server.Now().UTC()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = now
	}
	if request.ExpiresAt.IsZero() {
		request.ExpiresAt = now.Add(2 * time.Minute)
	}
	if err := request.Validate(now); err != nil {
		return nil, err
	}
	path, err := server.Queue.WritePending(request)
	if err != nil {
		return nil, err
	}
	return textResult(fmt.Sprintf("queued %s at %s", request.ID, path)), nil
}

func textResult(text string) map[string]any {
	return map[string]any{
		"isError": false,
		"content": []map[string]string{{
			"type": "text",
			"text": text,
		}},
	}
}
