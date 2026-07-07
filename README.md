# vmware-mcp-server

VMware Fusion MCP tooling for controlling a Windows VM from Codex.

Copyright (C) 2026 Burak Karahan. Licensed under LGPL-3.0-or-later.

## Components

- `vmware-mcp-server`: macOS host-side MCP stdio server for Codex.
- `vmware-mcp-agent`: Windows guest-side queue agent for launching apps and running commands.
- `vmware-uac-daemon`: macOS host-side UAC bridge daemon.
- `vmware-mcp-install`: installer helper for LaunchAgent and Windows service setup.

## Architecture

```text
Codex
  -> vmware-mcp-server
      -> vmrun
      -> shared queue
          -> vmware-mcp-agent.exe
              -> launch/run/status
              -> UAC request
          -> vmware-uac-daemon
              -> approve/reject UAC
```

The host MCP server owns VMware Fusion lifecycle operations. The Windows agent owns
guest OS behavior. UAC approval remains host-side because the guest cannot approve
its own secure desktop.

## Queue Layout

```text
queue/
  requests/
    <request-id>.pending.json
  responses/
    <request-id>.json
```

## Useful Environment

```sh
VMRUN_PATH="/Applications/VMware Fusion.app/Contents/Public/vmrun"
VMWARE_FUSION_DEFAULT_VMX="/Users/marlonjd/Virtual Machines.localized/Windows 11 64-bit Arm.vmwarevm/Windows 11 64-bit Arm.vmx"
VMWARE_FUSION_VM_PASSWORD_KEYCHAIN_SERVICE="codex-vmware-fusion-vm-password"
VMWARE_FUSION_VM_PASSWORD_KEYCHAIN_ACCOUNT="default"
VMWARE_FUSION_GUEST_USER="burak"
VMWARE_FUSION_GUEST_PASSWORD_KEYCHAIN_SERVICE="codex-vmware-fusion-guest-password"
VMWARE_FUSION_GUEST_PASSWORD_KEYCHAIN_ACCOUNT="default"
```

Do not store VM or guest passwords in this repository.

For UAC request signing, set the same shared secret on the Windows agent and
the macOS UAC daemon:

```sh
CODEX_UAC_BRIDGE_SECRET="<long-random-secret>"
```

The UAC daemon rejects signed-mode requests when the HMAC does not verify.

## Startup Daemons

Generate a macOS LaunchAgent plist for the UAC daemon:

```sh
vmware-mcp-install --mode launchagent \
  --binary /usr/local/bin/vmware-uac-daemon \
  --queue /path/to/shared/queue > ~/Library/LaunchAgents/com.burakkarahan.vmware-uac-daemon.plist
```

Generate PowerShell for a Windows startup task:

```powershell
vmware-mcp-install --mode windows-startup-task `
  --binary "C:\Program Files\vmware-mcp-tools\vmware-mcp-agent.exe" `
  --queue "\\vmware-host\Shared Folders\vmware-mcp-server\queue"
```

## Build

```sh
go test ./...
go build ./cmd/vmware-mcp-server
go build ./cmd/vmware-mcp-agent
go build ./cmd/vmware-uac-daemon
go build ./cmd/vmware-mcp-install
```
