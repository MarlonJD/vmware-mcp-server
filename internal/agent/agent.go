package agent

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

type Launcher interface {
	LaunchApp(context.Context, protocol.LaunchApp) error
}

type Runner interface {
	RunCommand(context.Context, protocol.RunCommand) error
}

type ExecLauncher struct{}

func (ExecLauncher) LaunchApp(ctx context.Context, launch protocol.LaunchApp) error {
	cmd := exec.CommandContext(ctx, launch.AppName, launch.Args...)
	if launch.CWD != "" {
		cmd.Dir = launch.CWD
	}
	return cmd.Start()
}

func (ExecLauncher) RunCommand(ctx context.Context, command protocol.RunCommand) error {
	cmd := exec.CommandContext(ctx, command.Command, command.Args...)
	if command.CWD != "" {
		cmd.Dir = command.CWD
	}
	return cmd.Start()
}

type Service struct {
	Launcher Launcher
	Runner   Runner
	Now      func() time.Time
}

func (service Service) Handle(ctx context.Context, request protocol.Request) protocol.Response {
	now := time.Now().UTC()
	if service.Now != nil {
		now = service.Now().UTC()
	}
	response := protocol.Response{RequestID: request.ID, HandledAt: now}

	if err := request.Validate(now); err != nil {
		response.Status = "rejected"
		response.Message = err.Error()
		return response
	}

	switch request.Kind {
	case protocol.RequestLaunchApp:
		launcher := service.Launcher
		if launcher == nil {
			launcher = ExecLauncher{}
		}
		if err := launcher.LaunchApp(ctx, *request.Launch); err != nil {
			response.Status = "failed"
			response.Message = err.Error()
			return response
		}
		response.Status = "completed"
		response.Message = fmt.Sprintf("launched %s", request.Launch.AppName)
	case protocol.RequestRunCommand:
		if request.Command.Admin {
			response.Status = "requires_uac"
			response.Message = "admin commands must be converted to signed UAC requests"
			return response
		}
		runner := service.Runner
		if runner == nil {
			runner = ExecLauncher{}
		}
		if err := runner.RunCommand(ctx, *request.Command); err != nil {
			response.Status = "failed"
			response.Message = err.Error()
			return response
		}
		response.Status = "completed"
		response.Message = fmt.Sprintf("started %s", request.Command.Command)
	default:
		response.Status = "rejected"
		response.Message = fmt.Sprintf("unsupported request kind: %s", request.Kind)
	}
	return response
}
