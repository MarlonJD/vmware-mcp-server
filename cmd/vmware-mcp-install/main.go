package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "help", "install target: launchagent, windows-startup-task, or help")
	binary := flag.String("binary", "/usr/local/bin/vmware-uac-daemon", "daemon binary path")
	queue := flag.String("queue", "/Users/marlonjd/Developer/demo/testvm/vmware-mcp-server/queue", "shared queue root")
	flag.Parse()

	switch *mode {
	case "launchagent":
		fmt.Print(launchAgentPlist(*binary, *queue))
	case "windows-startup-task", "windows-service":
		fmt.Print(windowsStartupTaskPowerShell(*binary, *queue))
	default:
		fmt.Fprintln(os.Stdout, "usage: vmware-mcp-install --mode launchagent|windows-startup-task")
	}
}

func launchAgentPlist(binary, queue string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.burakkarahan.vmware-uac-daemon</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>--queue</string>
    <string>%s</string>
    <string>--click</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
</dict>
</plist>
`, binary, queue)
}

func windowsStartupTaskPowerShell(binary, queue string) string {
	return fmt.Sprintf(
		"$binary = %q\n"+
			"$queue = %q\n"+
			"$action = New-ScheduledTaskAction -Execute $binary -Argument ('--queue \"' + $queue + '\"')\n"+
			"$trigger = New-ScheduledTaskTrigger -AtLogOn\n"+
			"$settings = New-ScheduledTaskSettingsSet -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)\n"+
			"Register-ScheduledTask -TaskName \"vmware-mcp-agent\" -Action $action -Trigger $trigger -Settings $settings -Description \"VMware MCP Agent\" -Force\n"+
			"Start-ScheduledTask -TaskName \"vmware-mcp-agent\"\n",
		binary,
		queue,
	)
}
