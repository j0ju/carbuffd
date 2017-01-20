// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"github.com/op/go-logging"
	"os"
	"path"
	"strings"
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
		backend   *logging.LogBackend
		logFormat logging.Formatter
	)

	if logFile == "-" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		logFormat = logging.MustStringFormatter(logFormatDefault)

	} else if logSyslog := strings.SplitN(logFile, ":", 2); logSyslog[0] == "syslog" {
		log.Fatalf("syslog not implemented")

	} else {
		if !path.IsAbs(logFile) {
			log.Fatalf("logfile must be an absolute path: '%s'", logFile)
		}
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		backend = logging.NewLogBackend(f, "", 0)
		logFormat = logging.MustStringFormatter(logFormatDefault)
	}

	formatted := logging.NewBackendFormatter(backend, logFormat)

	levelled := logging.AddModuleLevel(formatted)
	levelled.SetLevel(logging.Level(logLevel), "main")

	logging.SetBackend(levelled)
}

// vim: foldmethod=syntax
