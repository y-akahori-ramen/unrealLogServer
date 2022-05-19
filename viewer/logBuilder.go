package viewer

import (
	"strings"

	"github.com/y-akahori-ramen/unrealLogServer/db"
)

type Log struct {
	Log       string
	Category  string
	Verbosity string
}

type LogDataBuilder struct {
	logData []Log
}

func (l *LogDataBuilder) HandleLog(log db.LogData) error {
	verbosity := ToVerbosityNameForHTML(log.Verbosity)
	category := ToCategoryNameForHTML(log.Category)

	l.logData = append(l.logData, Log{Log: log.Log, Category: category, Verbosity: verbosity})
	return nil
}

func (l *LogDataBuilder) LogData() []Log {
	return l.logData
}

type LogStrBuilder struct {
	builder strings.Builder
}

func (l *LogStrBuilder) HandleLog(log db.LogData) error {
	_, err := l.builder.WriteString(log.Log)
	return err
}

func (l *LogStrBuilder) String() string {
	return l.builder.String()
}
