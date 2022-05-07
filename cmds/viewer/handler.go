package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
	"github.com/y-akahori-ramen/unrealLogServer/db"
)

type Handler struct {
	templates *template.Template
	querier   db.Querier
}

func NewHandler(querier db.Querier) (*Handler, error) {
	tmplates, err := template.ParseGlob("template/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{templates: tmplates, querier: querier}, nil
}

func (h *Handler) Renderer() echo.Renderer {
	return h
}

func (h *Handler) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return h.templates.ExecuteTemplate(w, name, data)
}

func getFileOpenAtStr(fileOpenAtUnixMilli int64) string {
	return time.UnixMilli(fileOpenAtUnixMilli).Format("2006/01/02 15:04:05")
}

func getLogIdStr(id unreallogserver.LogId) string {
	return fmt.Sprintf("%s_%s_%s", id.Host, id.Platform, getFileOpenAtStr(id.FileOpenAtUnixMilli))
}

func (h *Handler) HandleIndex(c echo.Context) error {
	const pageStep = 50
	pageStr := c.QueryParam("page")

	var curPage int
	if pageStr != "" {
		var err error
		curPage, err = strconv.Atoi(pageStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Parse page failed")
		}
	} else {
		curPage = 0
	}

	type LogInfo struct {
		Id         string
		FileOpenAt string
		Host       string
		Platform   string
		ViewerLink string
	}

	ids, err := h.querier.GetIds(c.Request().Context(), db.NewFilter(), curPage*pageStep, pageStep)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Query failed")
	}

	var logs []LogInfo
	for _, id := range ids {
		logs = append(logs, LogInfo{
			Id:         getLogIdStr(id),
			FileOpenAt: getFileOpenAtStr(id.FileOpenAtUnixMilli),
			Host:       id.Host,
			Platform:   id.Platform,
			ViewerLink: fmt.Sprintf("/viewer?host=%s&platform=%s&fileOpenAt=%d", id.Host, id.Platform, id.FileOpenAtUnixMilli),
		})
	}

	var nextPage int
	if len(ids) < pageStep {
		nextPage = -1
	} else {
		nextPage = curPage + 1
	}

	data := struct {
		Logs     []LogInfo
		PrevPage int
		NextPage int
	}{
		Logs:     logs,
		PrevPage: curPage - 1,
		NextPage: nextPage,
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func (h *Handler) HandleViewer(c echo.Context) error {
	host := c.QueryParam("host")
	platform := c.QueryParam("platform")
	fileOpenAtStr := c.QueryParam("fileOpenAt")
	if host == "" || platform == "" || fileOpenAtStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	var fileOpenAt int64
	if fileOpenAtStr != "" {
		var err error
		fileOpenAt, err = strconv.ParseInt(fileOpenAtStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Parse fileOpenAt failed")
		}
	}
	id := unreallogserver.LogId{Host: host, Platform: platform, FileOpenAtUnixMilli: fileOpenAt}

	logBuilder := LogBuilder{}
	err := h.querier.GetLog(c.Request().Context(), logBuilder.HandleLog, db.NewFilterFromLogID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Query failed")
	}
	log := logBuilder.String()
	if log == "" {
		log = "No Data"
	}

	data := struct {
		Log   string
		LogID string
	}{
		Log:   log,
		LogID: getLogIdStr(id),
	}

	return c.Render(http.StatusOK, "viewer.html", data)
}
