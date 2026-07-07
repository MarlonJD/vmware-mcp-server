package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/agent"
	"github.com/MarlonJD/vmware-mcp-server/internal/config"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
)

func main() {
	queueDir := flag.String("queue", config.QueueDir(), "shared queue root")
	once := flag.Bool("once", false, "process pending files once and exit")
	serviceMode := flag.Bool("service", false, "run as a Windows Service")
	interval := flag.Duration("interval", time.Second, "poll interval")
	flag.Parse()

	if *serviceMode {
		if err := runWindowsService(*queueDir, *interval); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	store := queue.New(*queueDir)
	service := agent.Service{}
	for {
		if err := agent.ProcessPending(context.Background(), store, service); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}
