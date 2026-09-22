package app

import (
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

const watchDebounce = 200 * time.Millisecond

// WatchMsg tells the model that the results tree changed.
type WatchMsg struct{}

// Watcher observes a results root recursively and reports coalesced changes.
type Watcher struct {
	fs     *fsnotify.Watcher
	send   func(tea.Msg)
	done   chan struct{}
	closed chan struct{}
}

// Watch starts watching root and all subdirectories. New directories are added
// as they appear so nested attempt artifacts are observed too.
func Watch(root string, send func(tea.Msg)) (*Watcher, error) {
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{fs: fs, send: send, done: make(chan struct{}), closed: make(chan struct{})}
	if err := watcher.addRecursive(root); err != nil {
		_ = fs.Close()
		return nil, err
	}
	go watcher.loop()
	return watcher, nil
}

// Close stops watching. It is safe to call more than once.
func (w *Watcher) Close() error {
	select {
	case <-w.done:
	default:
		close(w.done)
	}
	<-w.closed
	return w.fs.Close()
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return w.fs.Add(path)
		}
		return nil
	})
}

func (w *Watcher) loop() {
	defer close(w.closed)
	var timer *time.Timer
	var pending <-chan time.Time
	for {
		select {
		case event, ok := <-w.fs.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = w.addRecursive(event.Name)
				}
			}
			if timer == nil {
				timer = time.NewTimer(watchDebounce)
				pending = timer.C
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(watchDebounce)
			}
		case <-pending:
			pending = nil
			timer = nil
			w.send(WatchMsg{})
		case _, ok := <-w.fs.Errors:
			if !ok {
				return
			}
		case <-w.done:
			if timer != nil {
				timer.Stop()
			}
			return
		}
	}
}
