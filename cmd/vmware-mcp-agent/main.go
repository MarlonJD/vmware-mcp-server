package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/agent"
	"github.com/MarlonJD/vmware-mcp-server/internal/config"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
)

func main() {
	queueDir := flag.String("queue", config.QueueDir(), "shared queue root")
	once := flag.Bool("once", false, "process pending files once and exit")
	interval := flag.Duration("interval", time.Second, "poll interval")
	flag.Parse()

	store := queue.New(*queueDir)
	service := agent.Service{}
	for {
		if err := processPending(context.Background(), store, service); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}

func processPending(ctx context.Context, store queue.Store, service agent.Service) error {
	matches, err := filepath.Glob(filepath.Join(store.RequestsDir(), "*.pending.json"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		request, err := queue.ReadRequest(path)
		if err != nil {
			return err
		}
		if request.UAC != nil {
			continue
		}
		response := service.Handle(ctx, request)
		if _, err := store.WriteResponse(response); err != nil {
			return err
		}
		handledPath := strings.TrimSuffix(path, ".pending.json") + ".handled.json"
		_ = os.Rename(path, handledPath)
	}
	return nil
}
