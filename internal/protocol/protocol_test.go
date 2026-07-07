package protocol

import (
	"strings"
	"testing"
	"time"
)

func TestNewLaunchAppBuildsValidRequest(t *testing.T) {
	request := NewLaunchApp("Codex", []string{"--safe"})

	if request.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", request.SchemaVersion)
	}
	if !strings.HasPrefix(request.ID, "launch-") {
		t.Fatalf("id = %q, want launch prefix", request.ID)
	}
	if request.Kind != RequestLaunchApp {
		t.Fatalf("kind = %q, want launch_app", request.Kind)
	}
	if request.Launch.AppName != "Codex" {
		t.Fatalf("app = %q, want Codex", request.Launch.AppName)
	}
	if err := request.Validate(time.Now().UTC()); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsExpiredRequest(t *testing.T) {
	request := Request{
		SchemaVersion: 1,
		ID:            "launch-test",
		Kind:          RequestLaunchApp,
		CreatedAt:     time.Now().UTC().Add(-3 * time.Minute),
		ExpiresAt:     time.Now().UTC().Add(-time.Minute),
		Launch:        &LaunchApp{AppName: "Codex"},
	}

	if err := request.Validate(time.Now().UTC()); err == nil {
		t.Fatal("Validate returned nil, want expired error")
	}
}
