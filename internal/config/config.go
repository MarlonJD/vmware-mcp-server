package config

import (
	"os"
	"os/exec"

	"github.com/MarlonJD/vmware-mcp-server/internal/vmrun"
)

func HostVMRunConfig() vmrun.Config {
	return vmrun.Config{
		VMRunPath:     envDefault("VMRUN_PATH", "/Applications/VMware Fusion.app/Contents/Public/vmrun"),
		VMXPath:       os.Getenv("VMWARE_FUSION_DEFAULT_VMX"),
		VMPassword:    passwordFromEnvOrKeychain("VMWARE_FUSION_VM_PASSWORD", "VMWARE_FUSION_VM_PASSWORD_KEYCHAIN_SERVICE", "VMWARE_FUSION_VM_PASSWORD_KEYCHAIN_ACCOUNT"),
		GuestUser:     os.Getenv("VMWARE_FUSION_GUEST_USER"),
		GuestPassword: passwordFromEnvOrKeychain("VMWARE_FUSION_GUEST_PASSWORD", "VMWARE_FUSION_GUEST_PASSWORD_KEYCHAIN_SERVICE", "VMWARE_FUSION_GUEST_PASSWORD_KEYCHAIN_ACCOUNT"),
	}
}

func QueueDir() string {
	return envDefault("VMWARE_MCP_QUEUE_DIR", "queue")
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func passwordFromEnvOrKeychain(passwordEnv, serviceEnv, accountEnv string) string {
	if value := os.Getenv(passwordEnv); value != "" {
		return value
	}
	service := os.Getenv(serviceEnv)
	if service == "" {
		return ""
	}
	account := envDefault(accountEnv, "default")
	output, err := exec.Command("/usr/bin/security", "find-generic-password", "-w", "-s", service, "-a", account).Output()
	if err != nil {
		return ""
	}
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
	return string(output)
}
