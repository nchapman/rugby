// Package watcher provides file system watching for Rugby development.
package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event represents a file change event.
type Event struct {
	Path string
	Op   fsnotify.Op
}

// Watcher monitors files for changes and triggers rebuilds.
type Watcher struct {
	fsw      *fsnotify.Watcher
	root     string
	debounce time.Duration

	mu          sync.Mutex
	watchedDirs map[string]bool
}

// New creates a new file watcher for the given project root.
func New(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		fsw:         fsw,
		root:        root,
		debounce:    100 * time.Millisecond,
		watchedDirs: make(map[string]bool),
	}, nil
}

// SetDebounce sets the debounce duration for events.
// Must be called before Watch() to avoid data races.
func (w *Watcher) SetDebounce(d time.Duration) {
	w.debounce = d
}

// AddPath adds a file or directory to watch.
// If it's a directory, it watches recursively for .rg and .go files.
func (w *Watcher) AddPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return w.watchDirRecursive(absPath)
	}

	// For a file, watch its parent directory
	dir := filepath.Dir(absPath)
	return w.addDir(dir)
}

// watchDirRecursive adds a directory and all subdirectories to watch.
func (w *Watcher) watchDirRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories except .rugby (we don't watch that)
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return w.addDir(path)
		}
		return nil
	})
}

// addDir adds a single directory to the watch list.
func (w *Watcher) addDir(dir string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watchedDirs[dir] {
		return nil
	}

	if err := w.fsw.Add(dir); err != nil {
		return err
	}
	w.watchedDirs[dir] = true
	return nil
}

// Watch starts watching for file changes and sends debounced events to the channel.
// It blocks until the context is cancelled.
func (w *Watcher) Watch(ctx context.Context) (<-chan Event, <-chan error) {
	events := make(chan Event)
	errors := make(chan error, 1) // buffered to avoid blocking on errors

	go w.watchLoop(ctx, events, errors)

	return events, errors
}

// watchLoop is the main event loop that handles debouncing.
func (w *Watcher) watchLoop(ctx context.Context, events chan<- Event, errors chan<- error) {
	defer close(events)
	defer close(errors)

	var (
		timer   *time.Timer
		pending *Event
		timerMu sync.Mutex
	)

	// Clean up timer on exit
	defer func() {
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timerMu.Unlock()
	}()

	sendPending := func() {
		// Copy pending event under lock, then send outside lock to avoid deadlock
		timerMu.Lock()
		ev := pending
		pending = nil
		timerMu.Unlock()

		if ev != nil {
			select {
			case events <- *ev:
			case <-ctx.Done():
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			// Filter to only relevant files
			if !isRelevantFile(event.Name) {
				continue
			}

			// Handle new directories being created
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = w.addDir(event.Name)
				}
			}

			// Debounce: reset timer on each event
			timerMu.Lock()
			pending = &Event{Path: event.Name, Op: event.Op}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, sendPending)
			timerMu.Unlock()

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			select {
			case errors <- err:
			case <-ctx.Done():
				return
			default:
				// Drop error if channel is full to avoid blocking
			}
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	return w.fsw.Close()
}

// isRelevantFile returns true if the file should trigger a rebuild.
func isRelevantFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".rg" || ext == ".go"
}

// ClearScreen prints ANSI escape codes to clear the terminal.
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}
