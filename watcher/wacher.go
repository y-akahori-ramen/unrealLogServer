package watcher

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(\S+\s+\S+)`)
var fileOpenAtTimeLayout = "01/02/06 15:04:05"
var convertUTF8_LF = strings.NewReplacer(
	"\r\n", "\n",
	"\ufeff", "",
)

type Watcher struct {
	Logs          chan unreallogserver.Log
	handlerList   []unreallogserver.LogHandler
	fileOpenAt    time.Time
	watchInterval time.Duration
}

func NewWatcher(watchInterval time.Duration) *Watcher {
	wacher := &Watcher{Logs: make(chan unreallogserver.Log), watchInterval: watchInterval}
	return wacher
}

func (w *Watcher) AddHandler(handler unreallogserver.LogHandler) {
	w.handlerList = append(w.handlerList, handler)
}

func (w *Watcher) handleLog(log unreallogserver.Log) error {
	var err error
	err = nil

	//  UnrealEngineのログはUTF8WithBOMのCRLFで扱いにくいためUTF8のLFに変換する
	log.Log = convertUTF8_LF.Replace(log.Log)

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

	watcher := unreallognotify.NewWatcher(w.watchInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case log := <-watcher.Logs:
				noCategoryLog := log.Category == ""
				if noCategoryLog && logFileOpenPattern.MatchString(log.Log) {
					matches := logFileOpenPattern.FindStringSubmatch(log.Log)
					timeStr := matches[1]
					fileOpenAt, err := time.ParseInLocation(fileOpenAtTimeLayout, timeStr, time.Local)
					if err != nil {
						eventHandleResult <- err
						return
					}
					w.fileOpenAt = fileOpenAt
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
