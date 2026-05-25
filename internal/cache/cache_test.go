package cache

import (
	"path/filepath"
	"testing"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestGetSet(t *testing.T) {
	t.Run("empty store returns ok=false", func(t *testing.T) {
		s := New[int]("")
		if v, ok := s.Get(); ok || v != 0 {
			t.Fatalf("empty Get() = (%v, %v), want (0, false)", v, ok)
		}
	})

	t.Run("set then get returns value and ok=true", func(t *testing.T) {
		s := New[int]("")
		if err := s.Set(42); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		v, ok := s.Get()
		if !ok || v != 42 {
			t.Fatalf("Get() = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("set overwrites previous value", func(t *testing.T) {
		s := New[string]("")
		_ = s.Set("first")
		_ = s.Set("second")
		v, ok := s.Get()
		if !ok || v != "second" {
			t.Fatalf("Get() = (%q, %v), want (\"second\", true)", v, ok)
		}
	})
}

func TestDiskRoundtrip(t *testing.T) {
	dir := t.TempDir()
	// Nested path to exercise MkdirAll of parent directories.
	path := filepath.Join(dir, "nested", "cache.json")

	want := sample{Name: "htb", Count: 7}

	writer := New[sample](path)
	if err := writer.Set(want); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// A fresh store with the same path should restore the value.
	reader := New[sample](path)
	loaded, err := reader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !loaded {
		t.Fatal("Load() = false, want true")
	}

	got, ok := reader.Get()
	if !ok {
		t.Fatal("Get() ok = false after Load, want true")
	}
	if got != want {
		t.Fatalf("Get() = %+v, want %+v", got, want)
	}
}

func TestLoadNonexistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	s := New[sample](path)
	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if loaded {
		t.Fatal("Load() = true for nonexistent file, want false")
	}
	if _, ok := s.Get(); ok {
		t.Fatal("Get() ok = true after failed Load, want false")
	}
}

func TestLoadNoDiskPath(t *testing.T) {
	s := New[int]("")
	loaded, err := s.Load()
	if err != nil || loaded {
		t.Fatalf("Load() with no diskPath = (%v, %v), want (false, nil)", loaded, err)
	}
}
