package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
	"github.com/y-akahori-ramen/unrealLogServer/logger"
)

type TargetConfig struct {
	Tag      string `json:"tag"`
	Path     string `json:"path"`
	Platform string `json:"platform"`
}

type Target struct {
	config         TargetConfig
	fluentdHandler *logger.FluentdLogHandle
	logger         *logger.Logger
}

func NewTarget(config TargetConfig, fluentConf fluent.Config) (*Target, error) {
	if config.Tag == "" {
		return nil, fmt.Errorf("tag is none")
	}
	if config.Platform == "" {
		return nil, fmt.Errorf("patform is none")
	}
	if config.Path == "" {
		return nil, fmt.Errorf("path is none")
	}

	fluentdHandler, err := logger.NewFluentdLogHandle(config.Tag, config.Platform, fluentConf, time.Local)
	if err != nil {
		return nil, err
	}

	return &Target{fluentdHandler: fluentdHandler, config: config, logger: logger.NewLogger()}, nil
}

func (t *Target) Close() {
	t.fluentdHandler.Close()
}

func (t *Target) Wach(ctx context.Context, watchInterval time.Duration) error {
	log.Printf("Start waching. Tag:%s Platform:%s Path:%s", t.config.Tag, t.config.Platform, t.config.Platform)

	t.logger.AddHandler(ueloghandler.NewWatcherLogHandler(func(l ueloghandler.WatcherLog) error {
		err := t.fluentdHandler.HandleLog(l)
		if err != nil {
			return err
		}

		// Elastic searchのtimesampがログごとに分かれるようにelastic serachのtimestampの最小単位分スリープさせる
		time.Sleep(time.Millisecond)
		return nil
	}))

	return t.logger.Wach(ctx, t.config.Path, watchInterval)
}
