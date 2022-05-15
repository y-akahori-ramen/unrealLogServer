package watcher

import (
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"
	unreallogserver "github.com/y-akahori-ramen/unrealLogServer"
)

// FluentdLogHandle Send log to fluentd
type FluentdLogHandle struct {
	logger   *fluent.Fluent
	hostName string
	platform string
	tag      string
}

func NewFluentdLogHandle(tag string, platform string, fluentConf fluent.Config) (*FluentdLogHandle, error) {
	logger, err := fluent.New(fluent.Config(fluentConf))
	if err != nil {
		return &FluentdLogHandle{}, err
	}

	host, err := os.Hostname()
	if err != nil {
		return &FluentdLogHandle{}, err
	}
	return &FluentdLogHandle{platform: platform, tag: tag, hostName: host, logger: logger}, nil
}

func (h *FluentdLogHandle) Close() error {
	return h.logger.Close()
}

func (h *FluentdLogHandle) HandleLog(log unreallogserver.Log) error {
	logID := unreallogserver.LogId{Host: h.hostName, Platform: h.platform, FileOpenAtUnixMilli: log.FileOpenAt.UnixMilli()}
	logData := map[string]interface{}{
		"Host":                h.hostName,
		"Platform":            h.platform,
		"FileOpenAtUnixMilli": log.FileOpenAt.UnixMilli(),
		"Frame":               log.Frame,
		"Log":                 log.Log,
		"Category":            log.Category,
		"Verbosity":           log.Verbosity,
		"LogID":               logID.String(),
	}

	err := h.logger.Post(h.tag, logData)
	if err != nil {
		return err
	}

	return nil
}
