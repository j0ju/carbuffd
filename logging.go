// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"github.com/op/go-logging"
	"log/syslog"
	"os"
	"path"
)

const (
	logFormatDefault = "%{time:2006-01-02 15:04:05.000} %{level:8s} %{shortfunc}: %{message}"
	logFormatSyslog  = "%{shortfunc}: %{message}"
	logFormatPlain   = "%{level:8s} %{shortfunc}: %{message}"
)

var (
	log = logging.MustGetLogger("main")
)

func initLogging() {
	var (
		backend   logging.Backend
		logFormat logging.Formatter
		err       error
	)

	if Cfg.LogDestination == "-" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		logFormat = logging.MustStringFormatter(logFormatDefault)

	} else if Cfg.LogDestination == "systemd" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		logFormat = logging.MustStringFormatter(logFormatPlain)

	} else if Cfg.LogDestination == "syslog" {
		syslogPrefix := path.Base(os.Args[0])
		backend, err = logging.NewSyslogBackendPriority(syslogPrefix, syslog.LOG_LOCAL4)
		if err != nil {
			log.Fatalf("error opening channel to syslog %v", err)
		}
		logFormat = logging.MustStringFormatter(logFormatSyslog)

	} else {
		if !path.IsAbs(Cfg.LogDestination) {
			log.Fatalf("logfile must be an absolute path: '%s'", Cfg.LogDestination)
		}
		f, err := os.OpenFile(Cfg.LogDestination, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		backend = logging.NewLogBackend(f, "", 0)
		logFormat = logging.MustStringFormatter(logFormatDefault)
	}

	formatted := logging.NewBackendFormatter(backend, logFormat)

	levelled := logging.AddModuleLevel(formatted)
	levelled.SetLevel(logging.Level(Cfg.LogLevel), "main")

	logging.SetBackend(levelled)
}

// vim: foldmethod=syntax
