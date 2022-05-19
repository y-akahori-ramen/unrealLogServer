package logger

import (
	"os"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
	ueloghandler "github.com/y-akahori-ramen/ueLogHandler"
	"github.com/y-akahori-ramen/unrealLogServer/db"
)

// FluentdLogHandle Send log to fluentd
type FluentdLogHandle struct {
	logger   *fluent.Fluent
	hostName string
	platform string
	tag      string
	loc      *time.Location
}

func NewFluentdLogHandle(tag string, platform string, fluentConf fluent.Config, loc *time.Location) (*FluentdLogHandle, error) {
	logger, err := fluent.New(fluent.Config(fluentConf))
	if err != nil {
		return &FluentdLogHandle{}, err
	}

	host, err := os.Hostname()
	if err != nil {
		return &FluentdLogHandle{}, err
	}
	return &FluentdLogHandle{platform: platform, tag: tag, hostName: host, logger: logger, loc: loc}, nil
}

func (h *FluentdLogHandle) Close() error {
	return h.logger.Close()
}

func (h *FluentdLogHandle) HandleLog(log ueloghandler.Log) error {
	fileOpenTime, err := log.ParseFileOpenTime(h.loc)
	if err != nil {
		return err
	}

	logID := db.LogId{Host: h.hostName, Platform: h.platform, FileOpenAtUnixMilli: fileOpenTime.UnixMilli()}
	logData := map[string]interface{}{
		"Host":                h.hostName,
		"Platform":            h.platform,
		"FileOpenAtUnixMilli": fileOpenTime.UnixMilli(),
		"Frame":               log.Frame,
		"Log":                 log.Log,
		"Category":            log.Category,
		"Verbosity":           log.Verbosity,
		"LogID":               logID.String(),
	}

	err = h.logger.Post(h.tag, logData)
	if err != nil {
		return err
	}

	return nil
}
