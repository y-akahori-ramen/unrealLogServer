package viewer

import (
	"strings"

	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

type Log struct {
	Log       string
	Category  string
	Verbosity string
}

type LogBuilder struct {
	builder strings.Builder
	LogData []Log
}

func (l *LogBuilder) HandleLog(log unreallogserver.Log) error {
	verbosity := log.Verbosity
	if verbosity == "" {
		verbosity = "Log"
	}
	category := log.Category
	if category == "" {
		category = "(none)"
	}

	l.LogData = append(l.LogData, Log{Log: log.Log, Category: category, Verbosity: verbosity})
	_, err := l.builder.WriteString(log.Log)
	return err
}

func (l *LogBuilder) String() string {
	return l.builder.String()
}
