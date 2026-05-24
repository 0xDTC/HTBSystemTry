package cache

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestGetSet(t *testing.T) {
	t.Run("empty store returns ok=false", func(t *testing.T) {
		s := New[int](time.Minute, "")
		if v, ok := s.Get(); ok || v != 0 {
			t.Fatalf("empty Get() = (%v, %v), want (0, false)", v, ok)
		}
	})

	t.Run("set then get returns value and ok=true", func(t *testing.T) {
		s := New[int](time.Minute, "")
		if err := s.Set(42); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		v, ok := s.Get()
		if !ok || v != 42 {
			t.Fatalf("Get() = (%v, %v), want (42, true)", v, ok)
		}
	})

	t.Run("set overwrites previous value", func(t *testing.T) {
		s := New[string](time.Minute, "")
		_ = s.Set("first")
		_ = s.Set("second")
		v, ok := s.Get()
		if !ok || v != "second" {
			t.Fatalf("Get() = (%q, %v), want (\"second\", true)", v, ok)
		}
	})
}

func TestFresh(t *testing.T) {
	t.Run("empty store is not fresh", func(t *testing.T) {
		s := New[int](time.Minute, "")
		if s.Fresh() {
			t.Fatal("Fresh() = true on empty store, want false")
		}
	})

	t.Run("fresh right after Set with long TTL", func(t *testing.T) {
		s := New[int](time.Hour, "")
		_ = s.Set(1)
		if !s.Fresh() {
			t.Fatal("Fresh() = false right after Set, want true")
		}
	})

	t.Run("stale after TTL elapses", func(t *testing.T) {
		s := New[int](10*time.Millisecond, "")
		_ = s.Set(1)
		if !s.Fresh() {
			t.Fatal("Fresh() = false immediately after Set, want true")
		}
		time.Sleep(25 * time.Millisecond)
		if s.Fresh() {
			t.Fatal("Fresh() = true after TTL elapsed, want false")
		}
		// Value is still present even though stale.
		if _, ok := s.Get(); !ok {
			t.Fatal("Get() ok = false after TTL, want true (stale value retained)")
		}
	})
}

func TestAge(t *testing.T) {
	t.Run("zero when empty", func(t *testing.T) {
		s := New[int](time.Minute, "")
		if age := s.Age(); age != 0 {
			t.Fatalf("Age() = %v on empty store, want 0", age)
		}
	})

	t.Run("nonzero and grows after Set", func(t *testing.T) {
		s := New[int](time.Minute, "")
		_ = s.Set(1)
		time.Sleep(15 * time.Millisecond)
		if age := s.Age(); age < 10*time.Millisecond {
			t.Fatalf("Age() = %v, want >= 10ms", age)
		}
	})
}

func TestDiskRoundtrip(t *testing.T) {
	dir := t.TempDir()
	// Use a nested path to exercise MkdirAll of parent directories.
	path := filepath.Join(dir, "nested", "cache.json")

	want := sample{Name: "htb", Count: 7}

	writer := New[sample](time.Hour, path)
	if err := writer.Set(want); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// A fresh store with the same path should restore the value.
	reader := New[sample](time.Hour, path)
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

	// fetchedAt is preserved, so a long TTL means the restored value is fresh.
	if !reader.Fresh() {
		t.Fatal("Fresh() = false after Load with long TTL, want true")
	}
}

func TestLoadPreservesStaleness(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	// Persist with a tiny TTL and let it age past the TTL.
	writer := New[int](10*time.Millisecond, path)
	if err := writer.Set(99); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	reader := New[int](10*time.Millisecond, path)
	loaded, err := reader.Load()
	if err != nil || !loaded {
		t.Fatalf("Load() = (%v, %v), want (true, nil)", loaded, err)
	}
	// Value present but not fresh because fetchedAt was preserved.
	if _, ok := reader.Get(); !ok {
		t.Fatal("Get() ok = false after Load, want true")
	}
	if reader.Fresh() {
		t.Fatal("Fresh() = true after loading aged value, want false")
	}
}

func TestLoadNonexistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	s := New[sample](time.Hour, path)
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
	s := New[int](time.Hour, "")
	loaded, err := s.Load()
	if err != nil || loaded {
		t.Fatalf("Load() with no diskPath = (%v, %v), want (false, nil)", loaded, err)
	}
}

func TestRevalidateStaleWhileRevalidate(t *testing.T) {
	t.Run("empty store triggers fetch and onFresh asynchronously", func(t *testing.T) {
		s := New[int](time.Hour, "")

		fresh := make(chan int, 1)
		v, haveStale := Revalidate[int](
			s,
			func() (int, error) { return 123, nil },
			func(val int) { fresh <- val },
			func(err error) { t.Errorf("unexpected onErr: %v", err) },
		)

		// No stale value available yet.
		if haveStale || v != 0 {
			t.Fatalf("Revalidate() = (%v, %v), want (0, false)", v, haveStale)
		}

		select {
		case got := <-fresh:
			if got != 123 {
				t.Fatalf("onFresh got %d, want 123", got)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for onFresh")
		}

		// Store should now be populated and fresh.
		if val, ok := s.Get(); !ok || val != 123 {
			t.Fatalf("Get() = (%v, %v), want (123, true)", val, ok)
		}
		if !s.Fresh() {
			t.Fatal("Fresh() = false after revalidation, want true")
		}
	})

	t.Run("returns stale value immediately and refreshes in background", func(t *testing.T) {
		s := New[int](10*time.Millisecond, "")
		_ = s.Set(1)
		time.Sleep(25 * time.Millisecond) // now stale

		fresh := make(chan int, 1)
		v, haveStale := Revalidate[int](
			s,
			func() (int, error) { return 2, nil },
			func(val int) { fresh <- val },
			nil,
		)

		// Stale value is returned immediately.
		if !haveStale || v != 1 {
			t.Fatalf("Revalidate() = (%v, %v), want (1, true)", v, haveStale)
		}

		select {
		case got := <-fresh:
			if got != 2 {
				t.Fatalf("onFresh got %d, want 2", got)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for background refresh")
		}
	})

	t.Run("fresh store does not trigger fetch", func(t *testing.T) {
		s := New[int](time.Hour, "")
		_ = s.Set(5)

		called := make(chan struct{}, 1)
		v, haveStale := Revalidate[int](
			s,
			func() (int, error) {
				called <- struct{}{}
				return 6, nil
			},
			nil,
			nil,
		)

		if !haveStale || v != 5 {
			t.Fatalf("Revalidate() = (%v, %v), want (5, true)", v, haveStale)
		}

		select {
		case <-called:
			t.Fatal("fetch was called for a fresh store, want no fetch")
		case <-time.After(100 * time.Millisecond):
			// Expected: fetch not invoked.
		}
	})

	t.Run("fetch error invokes onErr", func(t *testing.T) {
		s := New[int](time.Hour, "")
		wantErr := errors.New("boom")

		errCh := make(chan error, 1)
		Revalidate[int](
			s,
			func() (int, error) { return 0, wantErr },
			func(int) { t.Error("unexpected onFresh on error") },
			func(err error) { errCh <- err },
		)

		select {
		case got := <-errCh:
			if !errors.Is(got, wantErr) {
				t.Fatalf("onErr got %v, want %v", got, wantErr)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for onErr")
		}

		// Store remains empty after a failed fetch.
		if _, ok := s.Get(); ok {
			t.Fatal("Get() ok = true after failed fetch, want false")
		}
	})

	t.Run("nil callbacks are safe", func(t *testing.T) {
		s := New[int](time.Hour, "")
		// Should not panic with nil onFresh/onErr.
		_, _ = Revalidate[int](s, func() (int, error) { return 7, nil }, nil, nil)

		// Give the goroutine a moment to run and populate the store.
		deadline := time.After(2 * time.Second)
		for {
			if v, ok := s.Get(); ok && v == 7 {
				return
			}
			select {
			case <-deadline:
				t.Fatal("store not populated after Revalidate with nil callbacks")
			case <-time.After(5 * time.Millisecond):
			}
		}
	})
}
