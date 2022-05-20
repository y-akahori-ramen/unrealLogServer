package elastic

import (
	"context"
	"fmt"

	"github.com/y-akahori-ramen/unrealLogServer/db"
)

type ElasticSession struct {
	querier     *ElasticQuerier
	filter      db.Filter
	searchAfter int
}

func NewElasticSession(querier *ElasticQuerier, filter db.Filter) *ElasticSession {
	return &ElasticSession{querier: querier, filter: filter}
}

func (s *ElasticSession) GetLog(ctx context.Context, logHandler db.LogHandler) error {
	if s.querier == nil {
		return fmt.Errorf("ElasticSession is closed")
	}

	const step = 1000
	for {
		logCount, nextSearchAfter, err := s.querier.getLog(ctx, logHandler, s.filter, s.searchAfter, step)
		if err != nil {
			return err
		}
		if logCount < step {
			break
		}
		s.searchAfter = nextSearchAfter
	}

	return nil

}

func (s *ElasticSession) Close() error {
	s.querier = nil
	s.searchAfter = 0
	return nil
}
