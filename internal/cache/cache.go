// Package cache provides a small, dependency-free, concurrency-safe typed cache
// with optional JSON disk persistence. It lets callers show persisted data
// instantly on startup (Load) and update it after a fetch (Set).
package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Store is a concurrency-safe, single-value typed cache. The zero value is not
// usable; create one with New.
type Store[T any] struct {
	mu       sync.RWMutex
	diskPath string
	value    T
	present  bool
}

// diskEnvelope is the on-disk JSON representation of a cached value.
type diskEnvelope[T any] struct {
	Value T `json:"value"`
}

// New creates a store. If diskPath != "", Set also persists the value
// (atomically) to that JSON file and Load can restore it.
func New[T any](diskPath string) *Store[T] {
	return &Store[T]{diskPath: diskPath}
}

// Get returns the current value and whether any value is present.
func (s *Store[T]) Get() (value T, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value, s.present
}

// Set stores value in memory and, if diskPath is configured, persists it to disk
// atomically (write temp file + os.Rename). Disk errors are returned, but the
// in-memory value is always updated.
func (s *Store[T]) Set(value T) error {
	s.mu.Lock()
	s.value = value
	s.present = true
	diskPath := s.diskPath
	s.mu.Unlock()

	if diskPath == "" {
		return nil
	}
	return persist(diskPath, diskEnvelope[T]{Value: value})
}

// persist writes env to path atomically: it ensures the parent directory
// exists, writes to a temporary file, then renames it into place.
func persist[T any](path string, env diskEnvelope[T]) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	data, err := json.Marshal(env)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp) // best-effort cleanup
		return err
	}
	return nil
}

// Load restores a previously persisted value from disk into memory. Returns
// (true, nil) if a value was loaded, (false, nil) if the file does not exist,
// (false, err) on a real error.
func (s *Store[T]) Load() (bool, error) {
	s.mu.RLock()
	diskPath := s.diskPath
	s.mu.RUnlock()

	if diskPath == "" {
		return false, nil
	}

	data, err := os.ReadFile(diskPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	var env diskEnvelope[T]
	if err := json.Unmarshal(data, &env); err != nil {
		return false, err
	}

	s.mu.Lock()
	s.value = env.Value
	s.present = true
	s.mu.Unlock()

	return true, nil
}
