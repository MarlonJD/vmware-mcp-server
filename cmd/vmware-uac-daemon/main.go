package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MarlonJD/vmware-mcp-server/internal/config"
	"github.com/MarlonJD/vmware-mcp-server/internal/queue"
	"github.com/MarlonJD/vmware-mcp-server/internal/uacdaemon"
)

type values []string

func (items *values) String() string {
	return strings.Join(*items, ",")
}

func (items *values) Set(value string) error {
	*items = append(*items, value)
	return nil
}

func main() {
	var apps values = []string{"Windows PowerShell"}
	var publishers values = []string{"Microsoft Windows"}
	queueDir := flag.String("queue", config.QueueDir(), "shared queue root")
	once := flag.Bool("once", false, "process pending files once and exit")
	click := flag.Bool("click", false, "click the visible UAC yes button after allowlist checks")
	interval := flag.Duration("interval", time.Second, "poll interval")
	flag.Var(&apps, "allow-app", "allowlisted UAC app name")
	flag.Var(&publishers, "allow-publisher", "allowlisted UAC publisher")
	flag.Parse()

	var approver uacdaemon.Approver
	if *click {
		approver = uacdaemon.OSAScriptApprover{YesXRatio: 0.439, YesYRatio: 0.672, ActivateDelay: 400 * time.Millisecond}
	}

	store := queue.New(*queueDir)
	daemon := uacdaemon.Daemon{
		AllowedApps:       apps,
		AllowedPublishers: publishers,
		Secret:            os.Getenv("CODEX_UAC_BRIDGE_SECRET"),
		Approver:          approver,
	}
	for {
		if err := processPending(context.Background(), store, daemon); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}

func processPending(ctx context.Context, store queue.Store, daemon uacdaemon.Daemon) error {
	matches, err := filepath.Glob(filepath.Join(store.RequestsDir(), "*.pending.json"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		request, err := queue.ReadRequest(path)
		if err != nil {
			return err
		}
		if request.UAC == nil {
			continue
		}
		response := daemon.Handle(ctx, request)
		if _, err := store.WriteResponse(response); err != nil {
			return err
		}
		handledPath := strings.TrimSuffix(path, ".pending.json") + "." + response.Status + ".json"
		_ = os.Rename(path, handledPath)
	}
	return nil
}
