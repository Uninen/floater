package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type windowStore struct {
	path     string
	mu       sync.Mutex
	revision uint64
}

func (s *windowStore) load() ([]savedWindow, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var windows []savedWindow
	if err := json.Unmarshal(data, &windows); err != nil {
		return nil, err
	}
	ids := make(map[string]bool)
	for _, w := range windows {
		if w.ID == "" || ids[w.ID] {
			return nil, fmt.Errorf("saved windows have missing or duplicate IDs")
		}
		if _, err := normalizeURL(w.URL); err != nil {
			return nil, fmt.Errorf("%s: %w", w.Name, err)
		}
		ids[w.ID] = true
	}
	return windows, nil
}

func (s *windowStore) save(revision uint64, windows []savedWindow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Save jobs can acquire the lock out of order; older snapshots must never win.
	if revision < s.revision {
		return nil
	}
	data, err := json.MarshalIndent(windows, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".windows-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(append(data, '\n'))
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), s.path); err != nil {
		return err
	}
	s.revision = revision
	return nil
}
