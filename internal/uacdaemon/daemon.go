package uacdaemon

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

type Approver interface {
	Approve(context.Context, protocol.UACApproval) error
}

type Daemon struct {
	AllowedApps       []string
	AllowedPublishers []string
	Secret            string
	Approver          Approver
	Now               func() time.Time
}

func (daemon Daemon) Handle(ctx context.Context, request protocol.Request) protocol.Response {
	now := time.Now().UTC()
	if daemon.Now != nil {
		now = daemon.Now().UTC()
	}
	response := protocol.Response{RequestID: request.ID, HandledAt: now}
	if err := request.Validate(now); err != nil {
		response.Status = "rejected"
		response.Message = err.Error()
		return response
	}
	if request.Kind != protocol.RequestUAC {
		response.Status = "rejected"
		response.Message = fmt.Sprintf("unsupported request kind: %s", request.Kind)
		return response
	}
	if !slices.Contains(daemon.AllowedApps, request.UAC.ExpectedApp) {
		response.Status = "rejected"
		response.Message = "expected app is not allowlisted"
		return response
	}
	if !slices.Contains(daemon.AllowedPublishers, request.UAC.ExpectedPublisher) {
		response.Status = "rejected"
		response.Message = "expected publisher is not allowlisted"
		return response
	}
	if daemon.Secret != "" && !protocol.VerifyUAC(request, daemon.Secret) {
		response.Status = "rejected"
		response.Message = "uac signature verification failed"
		return response
	}
	if daemon.Approver != nil {
		if err := daemon.Approver.Approve(ctx, *request.UAC); err != nil {
			response.Status = "failed"
			response.Message = err.Error()
			return response
		}
	}
	response.Status = "approved"
	response.Message = "uac request approved"
	return response
}
