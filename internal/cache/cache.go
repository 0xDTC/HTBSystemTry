// Package cache provides a small, dependency-free, concurrency-safe typed
// cache with a TTL and optional JSON disk persistence. It is designed to
// support a stale-while-revalidate pattern: callers can show stale data
// instantly while a fresh value is fetched in the background.
package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store is a concurrency-safe, single-value typed cache. The zero value is not
// usable; create one with New.
type Store[T any] struct {
	mu        sync.RWMutex
	ttl       time.Duration
	diskPath  string
	value     T
	fetchedAt time.Time
	present   bool
}

// diskEnvelope is the on-disk JSON representation of a cached value.
type diskEnvelope[T any] struct {
	Value     T         `json:"value"`
	FetchedAt time.Time `json:"fetched_at"`
}

// New creates a store with the given TTL. If diskPath != "", Set also persists
// the value (atomically) to that JSON file and Load can restore it.
func New[T any](ttl time.Duration, diskPath string) *Store[T] {
	return &Store[T]{
		ttl:      ttl,
		diskPath: diskPath,
	}
}

// Get returns the current value and whether any value is present (fresh OR
// stale).
func (s *Store[T]) Get() (value T, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value, s.present
}

// Fresh reports whether a value is present and was set within the TTL.
func (s *Store[T]) Fresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.present {
		return false
	}
	return time.Since(s.fetchedAt) < s.ttl
}

// Age returns how long since the value was set (0 if none present).
func (s *Store[T]) Age() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.present {
		return 0
	}
	return time.Since(s.fetchedAt)
}

// Set stores value in memory (stamped now) and, if diskPath is configured,
// persists {Value,FetchedAt} to disk atomically (write temp file + os.Rename).
// Disk errors are returned but the in-memory value is always updated.
func (s *Store[T]) Set(value T) error {
	now := time.Now()

	s.mu.Lock()
	s.value = value
	s.fetchedAt = now
	s.present = true
	diskPath := s.diskPath
	s.mu.Unlock()

	if diskPath == "" {
		return nil
	}
	return persist(diskPath, diskEnvelope[T]{Value: value, FetchedAt: now})
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
		// Best-effort cleanup of the temp file on failure.
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Load restores a previously persisted value from disk into memory. Returns
// (true, nil) if a value was loaded, (false, nil) if the file does not exist,
// (false, err) on a real error. The restored fetchedAt is preserved so Fresh()
// is honored across restarts.
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
	s.fetchedAt = env.FetchedAt
	s.present = true
	s.mu.Unlock()

	return true, nil
}

// Revalidate returns the current cached value immediately (stale allowed) via
// the returned (value, haveStale). If the store is missing or not Fresh, it
// launches fetch() in a NEW goroutine; on success it calls store.Set(v) then
// onFresh(v); on error it calls onErr(err) (onFresh/onErr may be nil). It never
// blocks on fetch.
func Revalidate[T any](store *Store[T], fetch func() (T, error), onFresh func(T), onErr func(error)) (value T, haveStale bool) {
	value, haveStale = store.Get()

	if !store.Fresh() {
		go func() {
			v, err := fetch()
			if err != nil {
				if onErr != nil {
					onErr(err)
				}
				return
			}
			if err := store.Set(v); err != nil {
				// The in-memory value is updated regardless of disk
				// errors; surface the persistence failure to the caller.
				if onErr != nil {
					onErr(err)
				}
				return
			}
			if onFresh != nil {
				onFresh(v)
			}
		}()
	}

	return value, haveStale
}
