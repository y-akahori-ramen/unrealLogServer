package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/y-akahori-ramen/unrealLogServer/db"
	elasticdb "github.com/y-akahori-ramen/unrealLogServer/db/elastic"
)

type Config struct {
	elasticsearch.Config
	CACertPath string
	LogIndex   string
	Address    string
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

	var querier db.Querier
	querier, err = elasticdb.NewElasticQuerier(config.LogIndex, *elasticConfig)
	if err != nil {
		log.Fatal(err)
	}

	handle, err := NewHandler(querier)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Renderer = handle.Renderer()
	e.Use(middleware.Logger())
	e.Static("/", "static")

	e.GET("/", handle.HandleIndex)
	e.GET("/viewer", handle.HandleViewer)

	e.Logger.Fatal(e.Start(config.Address))
}
