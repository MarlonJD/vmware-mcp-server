package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type RequestKind string

const (
	RequestLaunchApp RequestKind = "launch_app"
	RequestRunCommand RequestKind = "run_command"
	RequestUAC       RequestKind = "uac_request"
	RequestStatus    RequestKind = "status"
)

type Request struct {
	SchemaVersion int          `json:"schema_version"`
	ID            string       `json:"request_id"`
	Kind          RequestKind  `json:"kind"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     time.Time    `json:"expires_at"`
	Launch        *LaunchApp   `json:"launch,omitempty"`
	Command       *RunCommand  `json:"command,omitempty"`
	UAC           *UACApproval `json:"uac,omitempty"`
}

type LaunchApp struct {
	AppName string   `json:"app_name"`
	Args    []string `json:"args,omitempty"`
	CWD     string   `json:"cwd,omitempty"`
}

type RunCommand struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	CWD     string   `json:"cwd,omitempty"`
	Admin   bool     `json:"admin,omitempty"`
}

type UACApproval struct {
	ExpectedApp       string `json:"expected_app"`
	ExpectedPublisher string `json:"expected_publisher"`
	Reason            string `json:"reason"`
	CommandSummary    string `json:"command_summary"`
	Signature         string `json:"signature,omitempty"`
	SignatureAlgo     string `json:"signature_algorithm,omitempty"`
}

type Response struct {
	RequestID string    `json:"request_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	HandledAt time.Time `json:"handled_at"`
}

func NewLaunchApp(appName string, args []string) Request {
	now := time.Now().UTC()
	return Request{
		SchemaVersion: 1,
		ID:            NewID("launch"),
		Kind:          RequestLaunchApp,
		CreatedAt:     now,
		ExpiresAt:     now.Add(2 * time.Minute),
		Launch:        &LaunchApp{AppName: appName, Args: args},
	}
}

func NewID(prefix string) string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(bytes[:]))
	}
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

func (request Request) Validate(now time.Time) error {
	if request.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version: %d", request.SchemaVersion)
	}
	if strings.TrimSpace(request.ID) == "" {
		return errors.New("request_id is required")
	}
	if request.ExpiresAt.Before(now.UTC()) {
		return errors.New("request expired")
	}

	switch request.Kind {
	case RequestLaunchApp:
		if request.Launch == nil || strings.TrimSpace(request.Launch.AppName) == "" {
			return errors.New("launch.app_name is required")
		}
	case RequestRunCommand:
		if request.Command == nil || strings.TrimSpace(request.Command.Command) == "" {
			return errors.New("command.command is required")
		}
	case RequestUAC:
		if request.UAC == nil {
			return errors.New("uac payload is required")
		}
		if strings.TrimSpace(request.UAC.ExpectedApp) == "" {
			return errors.New("uac.expected_app is required")
		}
		if strings.TrimSpace(request.UAC.ExpectedPublisher) == "" {
			return errors.New("uac.expected_publisher is required")
		}
	case RequestStatus:
		return nil
	default:
		return fmt.Errorf("unknown request kind: %s", request.Kind)
	}

	return nil
}
