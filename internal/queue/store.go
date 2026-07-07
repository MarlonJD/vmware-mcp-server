package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MarlonJD/vmware-mcp-server/internal/protocol"
)

type Store struct {
	Root string
}

func New(root string) Store {
	return Store{Root: root}
}

func (store Store) RequestsDir() string {
	return filepath.Join(store.Root, "requests")
}

func (store Store) ResponsesDir() string {
	return filepath.Join(store.Root, "responses")
}

func (store Store) Ensure() error {
	if err := os.MkdirAll(store.RequestsDir(), 0o755); err != nil {
		return err
	}
	return os.MkdirAll(store.ResponsesDir(), 0o755)
}

func (store Store) WritePending(request protocol.Request) (string, error) {
	if err := store.Ensure(); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}
	payload = append(payload, '\n')

	finalPath := filepath.Join(store.RequestsDir(), fmt.Sprintf("%s.pending.json", request.ID))
	tempPath := finalPath + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", err
	}
	return finalPath, nil
}

func ReadRequest(path string) (protocol.Request, error) {
	var request protocol.Request
	payload, err := os.ReadFile(path)
	if err != nil {
		return request, err
	}
	err = json.Unmarshal(payload, &request)
	return request, err
}

func (store Store) WriteResponse(response protocol.Response) (string, error) {
	if err := store.Ensure(); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}
	payload = append(payload, '\n')
	path := filepath.Join(store.ResponsesDir(), fmt.Sprintf("%s.json", response.RequestID))
	return path, os.WriteFile(path, payload, 0o644)
}
