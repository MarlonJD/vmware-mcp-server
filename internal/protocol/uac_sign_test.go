package protocol

import (
	"testing"
	"time"
)

func TestSignAndVerifyUACRequest(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	request := Request{
		SchemaVersion: 1,
		ID:            "uac-test",
		Kind:          RequestUAC,
		CreatedAt:     now,
		ExpiresAt:     now.Add(time.Minute),
		UAC: &UACApproval{
			ExpectedApp:       "Windows PowerShell",
			ExpectedPublisher: "Microsoft Windows",
			Reason:            "install tools",
			CommandSummary:    "powershell setup.ps1",
		},
	}

	if err := SignUAC(&request, "secret"); err != nil {
		t.Fatalf("SignUAC returned error: %v", err)
	}
	if request.UAC.SignatureAlgo != "hmac-sha256-v1" {
		t.Fatalf("signature algorithm = %q", request.UAC.SignatureAlgo)
	}
	if !VerifyUAC(request, "secret") {
		t.Fatal("VerifyUAC returned false, want true")
	}
	if VerifyUAC(request, "wrong") {
		t.Fatal("VerifyUAC returned true for wrong secret")
	}
}
