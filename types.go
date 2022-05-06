package unreallogserver

import (
	"time"

	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
)

type Log struct {
	unreallognotify.LogInfo
	FileOpenAt time.Time
}

type LogHandler func(Log) error
