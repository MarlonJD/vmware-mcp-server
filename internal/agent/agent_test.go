package agent

import (
	"context"
	"testing"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

type recordingLauncher struct {
	app     string
	command string
}

func (launcher *recordingLauncher) LaunchApp(ctx context.Context, launch protocol.LaunchApp) error {
	launcher.app = launch.AppName
	return nil
}

func (launcher *recordingLauncher) RunCommand(ctx context.Context, command protocol.RunCommand) error {
	launcher.command = command.Command
	return nil
}

func TestServiceLaunchesAppRequest(t *testing.T) {
	recorder := &recordingLauncher{}
	service := Service{
		Launcher: recorder,
		Now:      func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) },
	}
	request := protocol.NewLaunchApp("Codex", nil)

	response := service.Handle(context.Background(), request)

	if response.Status != "completed" {
		t.Fatalf("status = %q, want completed: %s", response.Status, response.Message)
	}
	if recorder.app != "Codex" {
		t.Fatalf("launched app = %q, want Codex", recorder.app)
	}
}

func TestServiceRejectsExpiredRequest(t *testing.T) {
	service := Service{Now: func() time.Time { return time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC) }}
	request := protocol.NewLaunchApp("Codex", nil)
	request.ExpiresAt = time.Date(2026, 7, 7, 11, 59, 0, 0, time.UTC)

	response := service.Handle(context.Background(), request)

	if response.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", response.Status)
	}
}

func TestServiceRunsCommandRequest(t *testing.T) {
	recorder := &recordingLauncher{}
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	service := Service{
		Runner: recorder,
		Now:    func() time.Time { return now },
	}
	request := protocol.Request{
		SchemaVersion: 1,
		ID:            "cmd-test",
		Kind:          protocol.RequestRunCommand,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		Command:        &protocol.RunCommand{Command: "powershell.exe"},
	}

	response := service.Handle(context.Background(), request)

	if response.Status != "completed" {
		t.Fatalf("status = %q, want completed: %s", response.Status, response.Message)
	}
	if recorder.command != "powershell.exe" {
		t.Fatalf("command = %q, want powershell.exe", recorder.command)
	}
}

func TestServiceDoesNotRunAdminCommandWithoutUACFlow(t *testing.T) {
	recorder := &recordingLauncher{}
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	service := Service{
		Runner: recorder,
		Now:    func() time.Time { return now },
	}
	request := protocol.Request{
		SchemaVersion: 1,
		ID:            "cmd-test",
		Kind:          protocol.RequestRunCommand,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		Command:        &protocol.RunCommand{Command: "powershell.exe", Admin: true},
	}

	response := service.Handle(context.Background(), request)

	if response.Status != "requires_uac" {
		t.Fatalf("status = %q, want requires_uac", response.Status)
	}
	if recorder.command != "" {
		t.Fatalf("admin command ran unexpectedly: %q", recorder.command)
	}
}
