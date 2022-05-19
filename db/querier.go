package db

import (
	"context"
	"fmt"
)

type LogData struct {
	Log       string
	Category  string
	Verbosity string
}

type LogHandler func(LogData) error

type LogId struct {
	Host                string
	Platform            string
	FileOpenAtUnixMilli int64
}

func (id *LogId) String() string {
	return fmt.Sprintf("%s_%s_%v", id.Host, id.Platform, id.FileOpenAtUnixMilli)
}

type Querier interface {
	GetLog(ctx context.Context, logHandler LogHandler, filter Filter) error
	GetHosts(ctx context.Context, filter Filter) ([]string, error)
	GetPlatforms(ctx context.Context, filter Filter) ([]string, error)
	GetCategories(ctx context.Context, id LogId) ([]string, error)
	GetVerbosities(ctx context.Context, id LogId) ([]string, error)
	GetIds(ctx context.Context, filter Filter, from int, size int) ([]LogId, error)
}
