//go:build !windows

package main

import (
	"errors"
	"time"
)

func runWindowsService(queueDir string, interval time.Duration) error {
	return errors.New("--service is only supported on Windows")
}
