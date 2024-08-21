package utils

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"time"
)

const defaultInterval = 5 * time.Second

type OnChangeCallback func(path string)

type FileWatcher struct {
	filePath string
	info     fs.FileInfo
	exit     chan int
}

func NewFileWatcher(path string) (*FileWatcher, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	return &FileWatcher{filePath: path}, nil
}

func (w *FileWatcher) Watch(ctx context.Context, fn OnChangeCallback) {
	if w.exit != nil {
		w.Stop()
	}
	w.exit = make(chan int)
	w.info, _ = os.Stat(w.filePath)
	go w.checkIndefinitely(ctx, fn, defaultInterval)
}

func (w *FileWatcher) Stop() {
	w.exit <- 1
	close(w.exit)
}

func (w *FileWatcher) checkIndefinitely(ctx context.Context, fn OnChangeCallback, interval time.Duration) {

	for {
		select {
		case <-w.exit:
			return
		case <-ctx.Done():
			w.Stop()
			return
		default:
			fi, err := os.Stat(w.filePath)
			if err != nil {
				slog.Error("cannot access file. continue watching", "error", err)
			}
			if fi.Size() != w.info.Size() || fi.ModTime() != w.info.ModTime() {
				fn(w.filePath)
				w.info = fi
			}
		}
		time.Sleep(interval)
	}
}
