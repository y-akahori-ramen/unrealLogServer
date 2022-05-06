package unreallogserver

import (
	"fmt"
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

type Log struct {
	unreallognotify.LogInfo
	FileOpenAt time.Time
}

type LogHandler func(Log) error

type LogId struct {
	Host                string
	Platform            string
	FileOpenAtUnixMilli int64
}

func (id *LogId) String() string {
	return fmt.Sprintf("%s_%s_%v", id.Host, id.Platform, id.FileOpenAtUnixMilli)
}
