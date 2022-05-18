package logger

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

type Logger struct {
	w *ueloghandler.Watcher
}

func NewLogger() *Logger {
	return &Logger{w: ueloghandler.NewWatcher()}
}

func (l *Logger) AddHandler(handler ueloghandler.LogHandler) {
	l.w.AddHandler(handler)
}

func checkFileExist(ctx context.Context, filePath string) error {
	ticker := time.NewTicker(time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			_, err = os.Stat(filePath)
			if err == nil {
				log.Printf("File found. Path:%s", filePath)
				return nil
			} else if errors.Is(err, fs.ErrNotExist) {
				log.Printf("File not exist, retry after one scond. Path:%s", filePath)
			} else {
				return err
			}
		case <-ctx.Done():
			return err
		}
	}
}

func (l *Logger) Wach(ctx context.Context, filePath string, watchInterval time.Duration) error {
	for {
		err := checkFileExist(ctx, filePath)
		if err != nil {
			return err
		}

		err = l.w.Watch(ctx, filePath, watchInterval)
		if err != unreallognotify.ErrFileRemoved {
			return err
		}
	}
}
