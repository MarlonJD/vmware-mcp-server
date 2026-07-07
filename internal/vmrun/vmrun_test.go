package vmrun

import (
	"context"
	"reflect"
	"testing"
)

func TestRunProgramInGuestBuildsInteractiveCommand(t *testing.T) {
	var gotName string
	var gotArgs []string
	runner := Runner{
		Config: Config{
			VMRunPath:     "/opt/vmrun",
			VMXPath:       "/vms/win.vmx",
			VMPassword:    "vm-secret",
			GuestUser:     "burak",
			GuestPassword: "guest-secret",
		},
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return []byte("OK"), nil
		},
	}

	_, err := runner.RunProgramInGuest(
		context.Background(),
		"C:\\Windows\\System32\\notepad.exe",
		[]string{"a.txt"},
		GuestOptions{NoWait: true, ActiveWindow: true, Interactive: true},
	)
	if err != nil {
		t.Fatalf("RunProgramInGuest returned error: %v", err)
	}

	if gotName != "/opt/vmrun" {
		t.Fatalf("command name = %q", gotName)
	}
	want := []string{
		"-T", "fusion",
		"-vp", "vm-secret",
		"-gu", "burak",
		"-gp", "guest-secret",
		"runProgramInGuest",
		"/vms/win.vmx",
		"-noWait",
		"-activeWindow",
		"-interactive",
		"C:\\Windows\\System32\\notepad.exe",
		"a.txt",
	}
	if !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("args = %#v, want %#v", gotArgs, want)
	}
}
