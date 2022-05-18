package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

type Config struct {
	fluent.Config
	Targets               []TargetConfig `json:"targets"`
	WatchIntervalMilliSec int            `json:"watchintervalMilliSec"`
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

			results[idx] = target.Wach(ctx, time.Millisecond*time.Duration(config.WatchIntervalMilliSec))
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
