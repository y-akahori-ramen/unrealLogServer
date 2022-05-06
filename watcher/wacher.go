package watcher

import (
	"context"
	"regexp"
	"sync"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(.+)`)

type Watcher struct {
	Logs        chan unreallogserver.Log
	handlerList []unreallogserver.LogHandler
	fileOpenAt  string
}

func NewWatcher() *Watcher {
	wacher := &Watcher{Logs: make(chan unreallogserver.Log)}
	return wacher
}

func (w *Watcher) AddHandler(handler unreallogserver.LogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Watcher) handleLog(log unreallogserver.Log) error {
	var err error
	err = nil

	for _, handler := range w.handlerList {
		err = handler(log)
		if err != nil {
			return err
		}
	}
	return err
}

func (w *Watcher) Watch(ctx context.Context, filePath string) error {
	eventHandleResult := make(chan error)

	var wg sync.WaitGroup
	watchEnd := make(chan struct{})

	watcher := unreallognotify.NewWatcher()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case log := <-watcher.Logs:
				noCategoryLog := log.Category == ""
				if noCategoryLog && logFileOpenPattern.MatchString(log.Log) {
					matches := logFileOpenPattern.FindStringSubmatch(log.Log)
					w.fileOpenAt = matches[1]
				}

				logData := unreallogserver.Log{LogInfo: log, FileOpenAt: w.fileOpenAt}
				err := w.handleLog(logData)
				if err != nil {
					eventHandleResult <- err
					return
				}
			case <-watchEnd:
				return
			}
		}
	}()

	go func() {
		err := watcher.Watch(ctx, filePath)
		watcher.Flush()
		watchEnd <- struct{}{}
		eventHandleResult <- err
	}()

	err := <-eventHandleResult

	wg.Wait()
	return err
}
