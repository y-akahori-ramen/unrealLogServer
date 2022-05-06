package unreallogserver

import unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"

type Log struct {
	unreallognotify.LogInfo
	FileOpenAt string
}

type LogHandler func(Log) error
