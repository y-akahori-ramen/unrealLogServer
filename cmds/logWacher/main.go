package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	unreallognotify "github.com/y-akahori-ramen/unrealLogNotify"
	"github.com/y-akahori-ramen/unrealLogServer/watcher"
)

type TargetConfig struct {
	Tag      string `json:"tag"`
	Path     string `json:"path"`
	Platform string `json:"platform"`
}

type Config struct {
	fluent.Config
	Targets []TargetConfig `json:"targets"`
}

func LoadConfig(configPath string) (*Config, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Targets) == 0 {
		return nil, errors.New("targets is none")
	}

	return &config, nil
}

type Target struct {
	config         TargetConfig
	fluentdHandler *watcher.FluentdHandle
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

	fluentdHandler, err := watcher.NewFluentdHandle(config.Tag, config.Platform, fluentConf)
	if err != nil {
		return nil, err
	}

	return &Target{fluentdHandler: fluentdHandler, config: config}, nil
}

func (t *Target) Close() {
	t.fluentdHandler.Close()
}

func (t *Target) checkFileExist(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	var err error
	for {
		select {
		case <-ticker.C:
			_, err = os.Stat(t.config.Path)
			if err == nil {
				log.Printf("File found. Path:%s", t.config.Path)
				return nil
			} else if errors.Is(err, fs.ErrNotExist) {
				log.Printf("File not exist, retry after one scond. Path:%s", t.config.Path)
			} else {
				return err
			}
		case <-ctx.Done():
			return err
		}
	}
}

func (t *Target) Wach(ctx context.Context) error {
	log.Printf("Start waching. Tag:%s Platform:%s Path:%s", t.config.Tag, t.config.Platform, t.config.Platform)

	for {
		err := t.checkFileExist(ctx)
		if err != nil {
			return err
		}

		watcher := watcher.NewWatcher()
		watcher.AddHandler(t.handleLog)
		err = watcher.Watch(ctx, t.config.Path)
		if err != unreallognotify.ErrFileRemoved {
			return err
		}
	}
}

func (t *Target) handleLog(log watcher.Log) error {
	err := t.fluentdHandler.HandleLog(log)
	if err != nil {
		return err
	}

	// Elastic searchのtimesampがログごとに分かれるようにelastic serachのtimestampの最小単位分スリープさせる
	time.Sleep(time.Millisecond)
	return nil
}

func main() {
	configPath := flag.String("conf", "", "Path to config file.")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Load config error:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	results := make([]error, len(config.Targets))
	var wg sync.WaitGroup

	for targetIdx, target := range config.Targets {
		wg.Add(1)
		go func(idx int, targetConfig TargetConfig) {
			defer wg.Done()

			target, err := NewTarget(targetConfig, config.Config)
			if err != nil {
				results[idx] = err
				return
			}
			defer target.Close()

			results[idx] = target.Wach(ctx)
		}(targetIdx, target)
	}

	wg.Wait()

	hasError := false
	for targetIdx, err := range results {
		if err != nil {
			log.Printf("Target:%#v Error:%s", config.Targets[targetIdx], err.Error())
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
