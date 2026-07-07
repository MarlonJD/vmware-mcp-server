package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "help", "install target: launchagent, windows-service, or help")
	binary := flag.String("binary", "/usr/local/bin/vmware-uac-daemon", "daemon binary path")
	queue := flag.String("queue", "/Users/marlonjd/Developer/demo/testvm/vmware-mcp-server/queue", "shared queue root")
	flag.Parse()

	switch *mode {
	case "launchagent":
		fmt.Print(launchAgentPlist(*binary, *queue))
	case "windows-service":
		fmt.Print(windowsServicePowerShell(*binary, *queue))
	default:
		fmt.Fprintln(os.Stdout, "usage: vmware-mcp-install --mode launchagent|windows-service")
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

func windowsServicePowerShell(binary, queue string) string {
	return fmt.Sprintf(
		"$binary = %q\n"+
			"$queue = %q\n"+
			"$binPath = '\"' + $binary + '\" --service --queue \"' + $queue + '\"'\n"+
			"New-Service -Name \"vmware-mcp-agent\" -DisplayName \"VMware MCP Agent\" -StartupType Automatic -BinaryPathName $binPath\n"+
			"Start-Service -Name \"vmware-mcp-agent\"\n",
		binary,
		queue,
	)
}
