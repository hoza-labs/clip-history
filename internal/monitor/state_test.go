package monitor

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestSaveStateConcurrentReadersAndWriters verifies readers only observe complete state files.
func TestSaveStateConcurrentReadersAndWriters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "monitor.json")
	startedAt := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	if err := SaveState(path, State{PID: 1, StartedAt: startedAt, HeartbeatAt: startedAt}); err != nil {
		t.Fatalf("initial SaveState() error = %v", err)
	}

	const writerCount = 8
	const writesPerWriter = 50
	stopReaders := make(chan struct{})
	readerErrors := make(chan error, 1)
	var readers sync.WaitGroup
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stopReaders:
					return
				default:
					state, err := LoadState(path)
					if err != nil {
						select {
						case readerErrors <- err:
						default:
						}
						return
					}
					if state.PID <= 0 || state.StartedAt != startedAt || state.HeartbeatAt.IsZero() {
						select {
						case readerErrors <- errors.New("reader observed incomplete state"):
						default:
						}
						return
					}
				}
			}
		}()
	}

	var writers sync.WaitGroup
	for writer := range writerCount {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for write := range writesPerWriter {
				state := State{
					PID:         writer + 1,
					StartedAt:   startedAt,
					HeartbeatAt: startedAt.Add(time.Duration(write+1) * time.Nanosecond),
				}
				if err := SaveState(path, state); err != nil {
					select {
					case readerErrors <- err:
					default:
					}
					return
				}
			}
		}()
	}
	writers.Wait()
	close(stopReaders)
	readers.Wait()

	select {
	case err := <-readerErrors:
		t.Fatalf("concurrent state access error = %v", err)
	default:
	}

	temporaryFiles, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".monitor-state-*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary state files = %v, want none", temporaryFiles)
	}
}

// TestSaveStateReplacesExistingFile verifies a complete existing state is replaced.
func TestSaveStateReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "monitor.json")
	if err := os.WriteFile(path, []byte("old state"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	want := State{PID: 42, StartedAt: time.Now().UTC(), HeartbeatAt: time.Now().UTC()}
	if err := SaveState(path, want); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}
	got, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if got != want {
		t.Fatalf("LoadState() = %+v, want %+v", got, want)
	}
}
