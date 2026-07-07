package uacdaemon

import (
	"context"
	"testing"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

func TestDaemonApprovesAllowlistedUACRequest(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	daemon := Daemon{
		AllowedApps:       []string{"Windows PowerShell"},
		AllowedPublishers: []string{"Microsoft Windows"},
		Now:               func() time.Time { return now },
	}
	request := protocol.Request{
		SchemaVersion: 1,
		ID:            "uac-test",
		Kind:          protocol.RequestUAC,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		UAC: &protocol.UACApproval{
			ExpectedApp:       "Windows PowerShell",
			ExpectedPublisher: "Microsoft Windows",
			Reason:            "install dev tools",
		},
	}

	response := daemon.Handle(context.Background(), request)

	if response.Status != "approved" {
		t.Fatalf("status = %q, want approved: %s", response.Status, response.Message)
	}
}

func TestDaemonRejectsUnexpectedPublisher(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	daemon := Daemon{
		AllowedApps:       []string{"Windows PowerShell"},
		AllowedPublishers: []string{"Microsoft Windows"},
		Now:               func() time.Time { return now },
	}
	request := protocol.Request{
		SchemaVersion: 1,
		ID:            "uac-test",
		Kind:          protocol.RequestUAC,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		UAC: &protocol.UACApproval{
			ExpectedApp:       "Windows PowerShell",
			ExpectedPublisher: "Unknown Publisher",
		},
	}

	response := daemon.Handle(context.Background(), request)

	if response.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", response.Status)
	}
}

func TestDaemonRejectsBadSignatureWhenSecretConfigured(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	daemon := Daemon{
		AllowedApps:       []string{"Windows PowerShell"},
		AllowedPublishers: []string{"Microsoft Windows"},
		Secret:            "secret",
		Now:               func() time.Time { return now },
	}
	request := protocol.Request{
		SchemaVersion: 1,
		ID:            "uac-test",
		Kind:          protocol.RequestUAC,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		UAC: &protocol.UACApproval{
			ExpectedApp:       "Windows PowerShell",
			ExpectedPublisher: "Microsoft Windows",
			SignatureAlgo:     protocol.UACSignatureAlgorithm,
			Signature:         "bad",
		},
	}

	response := daemon.Handle(context.Background(), request)

	if response.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", response.Status)
	}
}
