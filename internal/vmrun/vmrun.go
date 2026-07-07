package vmrun

import (
	"context"
	"os/exec"
)

type Config struct {
	VMRunPath     string
	VMXPath       string
	VMPassword    string
	GuestUser     string
	GuestPassword string
}

type Runner struct {
	Config Config
	Run    func(context.Context, string, ...string) ([]byte, error)
}

type GuestOptions struct {
	NoWait       bool
	ActiveWindow bool
	Interactive  bool
}

func (runner Runner) command(ctx context.Context, args ...string) ([]byte, error) {
	if runner.Run != nil {
		return runner.Run(ctx, runner.Config.VMRunPath, args...)
	}
	return exec.CommandContext(ctx, runner.Config.VMRunPath, args...).CombinedOutput()
}

func (runner Runner) baseArgs() []string {
	args := []string{"-T", "fusion"}
	if runner.Config.VMPassword != "" {
		args = append(args, "-vp", runner.Config.VMPassword)
	}
	return args
}

func (runner Runner) guestArgs() []string {
	args := runner.baseArgs()
	if runner.Config.GuestUser != "" || runner.Config.GuestPassword != "" {
		args = append(args, "-gu", runner.Config.GuestUser, "-gp", runner.Config.GuestPassword)
	}
	return args
}

func (runner Runner) Start(ctx context.Context, gui bool) ([]byte, error) {
	mode := "nogui"
	if gui {
		mode = "gui"
	}
	args := append(runner.baseArgs(), "start", runner.Config.VMXPath, mode)
	return runner.command(ctx, args...)
}

func (runner Runner) GetGuestIP(ctx context.Context, wait bool) ([]byte, error) {
	args := append(runner.baseArgs(), "getGuestIPAddress", runner.Config.VMXPath)
	if wait {
		args = append(args, "-wait")
	}
	return runner.command(ctx, args...)
}

func (runner Runner) RunProgramInGuest(ctx context.Context, program string, programArgs []string, options GuestOptions) ([]byte, error) {
	args := append(runner.guestArgs(), "runProgramInGuest", runner.Config.VMXPath)
	if options.NoWait {
		args = append(args, "-noWait")
	}
	if options.ActiveWindow {
		args = append(args, "-activeWindow")
	}
	if options.Interactive {
		args = append(args, "-interactive")
	}
	args = append(args, program)
	args = append(args, programArgs...)
	return runner.command(ctx, args...)
}
