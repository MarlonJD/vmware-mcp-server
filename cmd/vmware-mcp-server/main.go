package main

import (
	"context"
	"fmt"
	"os"

	"github.com/MarlonJD/vmware-mcp-server/internal/config"
	"github.com/MarlonJD/vmware-mcp-server/internal/mcp"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
	"github.com/MarlonJD/vmware-mcp-server/internal/stdio"
	"github.com/MarlonJD/vmware-mcp-server/internal/vmrun"
)

func main() {
	server := mcp.Server{
		Queue: queue.New(config.QueueDir()),
		VMRun: vmrun.Runner{
			Config: config.HostVMRunConfig(),
		},
	}
	err := stdio.ReadMessages(os.Stdin, func(payload []byte) ([]byte, error) {
		return server.Handle(context.Background(), payload)
	}, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
