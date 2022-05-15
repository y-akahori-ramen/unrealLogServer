package viewer

import (
	"context"
	"embed"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/y-akahori-ramen/unrealLogServer/db"
)

type Server struct {
	e       *echo.Echo
	handler *Handler
}

//go:embed static/*
var staticAssets embed.FS

func NewServer(querier db.Querier, timeLocation *time.Location) (*Server, error) {
	handler, err := NewHandler(querier, timeLocation)
	if err != nil {
		return nil, err
	}

	e := echo.New()
	e.Renderer = handler.Renderer()
	e.Use(middleware.Logger())
	e.StaticFS("/", staticAssets)

	e.GET("/", handler.HandleIndex)
	e.GET("/viewer", handler.HandleViewer)
	e.GET("/download", handler.HandleDownloadLog)

	return &Server{e: e, handler: handler}, nil
}

func (s *Server) Start(address string) error {
	return s.e.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
