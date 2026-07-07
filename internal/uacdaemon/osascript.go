package uacdaemon

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

type OSAScriptApprover struct {
	YesXRatio     float64
	YesYRatio     float64
	ActivateDelay time.Duration
}

func (approver OSAScriptApprover) Approve(ctx context.Context, approval protocol.UACApproval) error {
	delaySeconds := approver.ActivateDelay.Seconds()
	if delaySeconds == 0 {
		delaySeconds = 0.4
	}
	xRatio := approver.YesXRatio
	if xRatio == 0 {
		xRatio = 0.439
	}
	yRatio := approver.YesYRatio
	if yRatio == 0 {
		yRatio = 0.672
	}
	script := fmt.Sprintf(`
tell application "System Events" to tell process "VMware Fusion" to set frontmost to true
delay %.3f
tell application "System Events"
  tell process "VMware Fusion"
    set vmWindow to front window
    set windowPosition to position of vmWindow
    set windowSize to size of vmWindow
    set clickX to (item 1 of windowPosition) + ((item 1 of windowSize) * %.6f)
    set clickY to (item 2 of windowPosition) + ((item 2 of windowSize) * %.6f)
  end tell
  click at {clickX as integer, clickY as integer}
end tell
`, delaySeconds, xRatio, yRatio)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}
