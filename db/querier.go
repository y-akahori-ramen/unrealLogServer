package db

import (
	"context"

	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

type Querier interface {
	GetLog(ctx context.Context, logHandler unreallogserver.LogHandler, filter Filter) error
	GetHosts(ctx context.Context, filter Filter) ([]string, error)
	GetPlatforms(ctx context.Context, filter Filter) ([]string, error)
	GetCategories(ctx context.Context, id unreallogserver.LogId) ([]string, error)
	GetVerbosities(ctx context.Context, id unreallogserver.LogId) ([]string, error)
	GetIds(ctx context.Context, filter Filter, from int, size int) ([]unreallogserver.LogId, error)
}
