package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSONAtomic(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(path)
		if secondErr := os.Rename(tmp, path); secondErr != nil {
			return fmt.Errorf("rename temp json: %v / %v", err, secondErr)
		}
	}
	return nil
}
