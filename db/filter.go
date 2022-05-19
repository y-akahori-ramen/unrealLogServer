package db

import (
	"time"
)

type Verbosity uint32

const (
	Log Verbosity = 1 << iota
	Warning
	Error
	Display
	Verbose
	VeryVerbose
	VerbosityNum = iota
	None         = 0
)

type Filter struct {
	Verbosity           Verbosity
	Categories          []string
	Hosts               []string
	Platforms           []string
	FileOpenAtUnixMilli int64
	timeFileterEnable   bool
	from                time.Time
	to                  time.Time
}

func NewFilter() Filter {
	return Filter{}
}

func NewFilterFromLogID(id LogId) Filter {
	return Filter{Hosts: []string{id.Host}, Platforms: []string{id.Platform}, FileOpenAtUnixMilli: id.FileOpenAtUnixMilli}
}

func (f *Filter) SetTimeRange(from, to time.Time) {
	f.timeFileterEnable = true
	f.from = from
	f.to = to
}

func (f *Filter) GetTimeRange() (time.Time, time.Time, bool) {
	return f.from, f.to, f.timeFileterEnable
}
