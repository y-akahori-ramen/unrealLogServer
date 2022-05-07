package main

import (
	"strings"

	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

type LogBuilder struct {
	builder strings.Builder
}

func (l *LogBuilder) HandleLog(log unreallogserver.Log) error {
	_, err := l.builder.WriteString(log.Log)
	return err
}

func (l *LogBuilder) String() string {
	return l.builder.String()
}
