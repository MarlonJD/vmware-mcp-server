package main

import "testing"

func TestWindowsServicePowerShellInstallsRealService(t *testing.T) {
	script := windowsServicePowerShell(
		`C:\Program Files\vmware-mcp-tools\vmware-mcp-agent.exe`,
		`\\vmware-host\Shared Folders\vmware-mcp-server\queue`,
	)

	for _, want := range []string{
		"New-Service",
		"Start-Service",
		"--service",
		"vmware-mcp-agent",
	} {
		if !contains(script, want) {
			t.Fatalf("script missing %q:\n%s", want, script)
		}
	}
}

func contains(text, want string) bool {
	return len(want) == 0 || (len(text) >= len(want) && index(text, want) >= 0)
}

func index(text, want string) int {
	for i := 0; i+len(want) <= len(text); i++ {
		if text[i:i+len(want)] == want {
			return i
		}
	}
	return -1
}
