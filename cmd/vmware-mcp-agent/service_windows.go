//go:build windows

package main

import (
	"context"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/agent"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
	"golang.org/x/sys/windows/svc"
)

const serviceName = "vmware-mcp-agent"

func runWindowsService(queueDir string, interval time.Duration) error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if !isService {
		store := queue.New(queueDir)
		service := agent.Service{}
		for {
			if err := agent.ProcessPending(context.Background(), store, service); err != nil {
				return err
			}
			time.Sleep(interval)
		}
	}
	return svc.Run(serviceName, serviceHandler{queueDir: queueDir, interval: interval})
}

type serviceHandler struct {
	queueDir string
	interval time.Duration
}

func (handler serviceHandler) Execute(args []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := queue.New(handler.queueDir)
	service := agent.Service{}
	ticker := time.NewTicker(handler.interval)
	defer ticker.Stop()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case <-ticker.C:
			_ = agent.ProcessPending(ctx, store, service)
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				changes <- request.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			default:
				changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
			}
		}
	}
}
