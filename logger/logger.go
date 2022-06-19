package logger

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"regexp"
	"time"

	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
)

var logFileOpenPattern = regexp.MustCompile(`Log\sfile\sopen,\s+(\S+\s+\S+)`)

type Log struct {
	LogData      ueloghandler.Log
	FileOpenTime string
}

func (l *Log) ParseFileOpenTime(loc *time.Location) (time.Time, error) {
	if l.FileOpenTime == "" {
		return time.Time{}, ueloghandler.ErrNoTimeData
	}

	const fileOpenTimeLayout = "01/02/06 15:04:05"
	return time.ParseInLocation(fileOpenTimeLayout, l.FileOpenTime, loc)
}

type LogHandler interface {
	HandleLog(log Log) error
}

func NewLogHandler(function func(log Log) error) LogHandler {
	return &funcLogHanlder{function: function}
}

type funcLogHanlder struct {
	function func(log Log) error
}

func (h *funcLogHanlder) HandleLog(log Log) error {
	return h.function(log)
}

type Logger struct {
	w            *ueloghandler.Watcher
	fileOpenTime string
	handlerList  []LogHandler
}

func NewLogger() *Logger {
	watcher := ueloghandler.NewWatcher()
	logger := &Logger{w: watcher}
	watcher.AddLogHandler(ueloghandler.NewLogHandler(logger.handleLog))
	return logger
}

func (l *Logger) AddHandler(handler LogHandler) {
	l.handlerList = append(l.handlerList, handler)
}

func (l *Logger) Wach(ctx context.Context, filePath string, watchInterval time.Duration) error {
	for {
		err := checkFileExist(ctx, filePath)
		if err != nil {
			return err
		}

		fileNotifler := ueloghandler.NewFileNotifier(filePath, watchInterval)
		err = l.w.Watch(ctx, fileNotifler)
		if err != ueloghandler.ErrFileRemoved {
			return err
		}
	}
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

func (l *Logger) handleLog(log ueloghandler.Log) error {

	if log.Category == "" && logFileOpenPattern.MatchString(log.Log) {
		matches := logFileOpenPattern.FindStringSubmatch(log.Log)
		l.fileOpenTime = matches[1]
	}

	return l.handleLoggerLog(Log{LogData: log, FileOpenTime: l.fileOpenTime})
}

func (l *Logger) handleLoggerLog(log Log) error {
	for _, handler := range l.handlerList {
		err := handler.HandleLog(log)
		if err != nil {
			return err
		}
	}
	return nil
}
