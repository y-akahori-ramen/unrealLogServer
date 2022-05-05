package watcher

import (
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"
)

type FluentdHandle struct {
	logger   *fluent.Fluent
	hostName string
	platform string
	tag      string
}

func NewFluentdHandle(tag string, platform string, fluentConf fluent.Config) (*FluentdHandle, error) {
	logger, err := fluent.New(fluent.Config(fluentConf))
	if err != nil {
		return &FluentdHandle{}, err
	}

	host, err := os.Hostname()
	if err != nil {
		return &FluentdHandle{}, err
	}
	return &FluentdHandle{platform: platform, tag: tag, hostName: host, logger: logger}, nil
}

func (h *FluentdHandle) Close() {
	h.logger.Close()
}

func (h *FluentdHandle) HandleLog(log Log) error {
	logData := map[string]string{
		"Host":       h.hostName,
		"Platform":   h.platform,
		"FileOpenAt": log.FileOpenAt,
		"Frame":      log.Frame,
		"Log":        log.Log,
		"Category":   log.Category,
		"Verbosity":  log.Verbosity,
	}

	err := h.logger.Post(h.tag, logData)
	if err != nil {
		return err
	}

	return nil
}
