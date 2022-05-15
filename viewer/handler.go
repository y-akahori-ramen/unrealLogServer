package viewer

import (
	"embed"
	"errors"
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
	templates    *template.Template
	querier      db.Querier
	timeLocation *time.Location
}

//go:embed template/*
var temlateAssets embed.FS

func NewHandler(querier db.Querier, timeLocation *time.Location) (*Handler, error) {
	tmplates, err := template.ParseFS(temlateAssets, "template/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{templates: tmplates, querier: querier, timeLocation: timeLocation}, nil
}

func (h *Handler) Renderer() echo.Renderer {
	return h
}

func (h *Handler) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return h.templates.ExecuteTemplate(w, name, data)
}

func (h *Handler) getFileOpenAtStr(fileOpenAtUnixMilli int64) string {
	return time.UnixMilli(fileOpenAtUnixMilli).In(h.timeLocation).Format("2006/01/02 15:04:05")
}

func (h *Handler) getLogIdStr(id unreallogserver.LogId) string {
	return fmt.Sprintf("%s_%s_%s", id.Host, id.Platform, h.getFileOpenAtStr(id.FileOpenAtUnixMilli))
}

func getLogIdQueryParam(id unreallogserver.LogId) string {
	return fmt.Sprintf("host=%s&platform=%s&fileOpenAt=%d", id.Host, id.Platform, id.FileOpenAtUnixMilli)
}

func getLogIdFromQuery(c echo.Context) (unreallogserver.LogId, error) {
	host := c.QueryParam("host")
	platform := c.QueryParam("platform")
	fileOpenAtStr := c.QueryParam("fileOpenAt")
	if host == "" || platform == "" || fileOpenAtStr == "" {
		return unreallogserver.LogId{}, errors.New("Invalid QueryParam")
	}

	var fileOpenAt int64
	if fileOpenAtStr != "" {
		var err error
		fileOpenAt, err = strconv.ParseInt(fileOpenAtStr, 10, 64)
		if err != nil {
			return unreallogserver.LogId{}, errors.New("Parse fileOpenAt failed")
		}
	}
	id := unreallogserver.LogId{Host: host, Platform: platform, FileOpenAtUnixMilli: fileOpenAt}
	return id, nil
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
		Id           string
		FileOpenAt   string
		Host         string
		Platform     string
		ViewerLink   string
		DownloadLink string
	}

	ids, err := h.querier.GetIds(c.Request().Context(), db.NewFilter(), curPage*pageStep, pageStep)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Query failed")
	}

	var logs []LogInfo
	for _, id := range ids {
		logs = append(logs, LogInfo{
			Id:           h.getLogIdStr(id),
			FileOpenAt:   h.getFileOpenAtStr(id.FileOpenAtUnixMilli),
			Host:         id.Host,
			Platform:     id.Platform,
			ViewerLink:   fmt.Sprintf("/viewer?%s", getLogIdQueryParam(id)),
			DownloadLink: fmt.Sprintf("/download?%s", getLogIdQueryParam(id)),
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

var verbosityNames = []string{
	"Log",
	"Warning",
	"Error",
	"Display",
	"Verbose",
	"VeryVerbose",
}

func ToCategoryNameForHTML(categoryName string) string {
	// カテゴリなしはDBには空文字で登録されているがHTML上で分かりにくいため(none)という文字列で表示する
	if categoryName == "" {
		return "(none)"
	} else {
		return categoryName
	}
}

func ToVerbosityNameForHTML(verbosityName string) string {
	// LogレベルはUE上ではVerbosity名が表示されないため""となっている。HTML上でわかりにくいため"Log"という文字列で表示する
	if verbosityName == "" {
		return "Log"
	} else {
		return verbosityName
	}
}

func (h *Handler) HandleViewer(c echo.Context) error {
	id, err := getLogIdFromQuery(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	categoryNames := []string{}
	categories, err := h.querier.GetCategories(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	for _, category := range categories {
		categoryNames = append(categoryNames, ToCategoryNameForHTML(category))
	}

	logBuilder := LogDataBuilder{}

	err = h.querier.GetLog(c.Request().Context(), logBuilder.HandleLog, db.NewFilterFromLogID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Query failed")
	}

	data := struct {
		LogID            string
		DownloadLink     string
		LogIdQuery       string
		VerbosityNames   []string
		CategoryJsonData []*CategoryData
		LogData          []Log
	}{
		LogID:            h.getLogIdStr(id),
		DownloadLink:     fmt.Sprintf("/download?%s", getLogIdQueryParam(id)),
		LogIdQuery:       getLogIdQueryParam(id),
		VerbosityNames:   verbosityNames,
		CategoryJsonData: []*CategoryData{NewCaregoryDataBuilder().CreateCategoryData(categoryNames)},
		LogData:          logBuilder.LogData(),
	}

	return c.Render(http.StatusOK, "viewer.html", data)
}

func (h *Handler) HandleDownloadLog(c echo.Context) error {
	id, err := getLogIdFromQuery(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	logBuilder := LogStrBuilder{}
	err = h.querier.GetLog(c.Request().Context(), logBuilder.HandleLog, db.NewFilterFromLogID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Query failed")
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s.log\"", h.getLogIdStr(id)))

	return c.Blob(http.StatusOK, "text/plain", []byte(logBuilder.String()))
}
