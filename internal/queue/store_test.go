package queue

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

func TestWritePendingUsesAtomicPendingFileName(t *testing.T) {
	store := New(t.TempDir())
	request := protocol.NewLaunchApp("Codex", nil)

	path, err := store.WritePending(request)
	if err != nil {
		t.Fatalf("WritePending returned error: %v", err)
	}

	want := filepath.Join(store.RequestsDir(), request.ID+".pending.json")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("pending file missing: %v", err)
	}
}

func TestReadRequestRoundTrip(t *testing.T) {
	store := New(t.TempDir())
	request := protocol.NewLaunchApp("Codex", []string{"--debug"})
	path, err := store.WritePending(request)
	if err != nil {
		t.Fatalf("WritePending returned error: %v", err)
	}

	got, err := ReadRequest(path)
	if err != nil {
		t.Fatalf("ReadRequest returned error: %v", err)
	}
	if got.ID != request.ID || got.Launch.AppName != "Codex" {
		t.Fatalf("request round trip mismatch: %#v", got)
	}
	if err := got.Validate(time.Now().UTC()); err != nil {
		t.Fatalf("round-tripped request does not validate: %v", err)
	}
}
