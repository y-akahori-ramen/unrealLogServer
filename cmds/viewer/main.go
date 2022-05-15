package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/y-akahori-ramen/unrealLogServer/db"
	elasticdb "github.com/y-akahori-ramen/unrealLogServer/db/elastic"
	"github.com/y-akahori-ramen/unrealLogServer/viewer"
)

type Config struct {
	elasticsearch.Config
	CACertPath   string
	LogIndex     string
	Address      string
	TimeLocation string
}

func (c *Config) CreateElasticConfig() (*elasticsearch.Config, error) {
	elasticConfig := c.Config

	if c.CACertPath != "" {
		cert, err := ioutil.ReadFile(c.CACertPath)
		if err != nil {
			return nil, err
		}
		elasticConfig.CACert = cert
	}

	return &elasticConfig, nil
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
	return &config, nil
}

func main() {
	configPath := flag.String("conf", "", "Path to config file.")
	flag.Parse()
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Load config error:", err)
	}
	elasticConfig, err := config.CreateElasticConfig()
	if err != nil {
		log.Fatal("Create elastic config error:", err)
	}

	var timeLocation *time.Location
	if config.TimeLocation != "" {
		timeLocation, err = time.LoadLocation("Asia/Tokyo")
		if err != nil {
			log.Fatal("Load time location error:", err)
		}
	} else {
		timeLocation = time.Local
	}

	var querier db.Querier
	querier, err = elasticdb.NewElasticQuerier(config.LogIndex, *elasticConfig)
	if err != nil {
		log.Fatal(err)
	}

	server, err := viewer.NewServer(querier, timeLocation)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(config.Address); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
