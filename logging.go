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
	// should go into avoidDuplicatesBackend
	lastLevel   logging.Level
	lastMessage string
)

type avoidDuplicatesBackend struct {
	backend logging.Backend
}

func (adbe avoidDuplicatesBackend) Log(l logging.Level, calldepth int, r *logging.Record) error {
	if r.Level == lastLevel && r.Message() == lastMessage {
		return nil
	}

	lastLevel = r.Level
	lastMessage = r.Message()

	calldepth += 2
	return adbe.backend.Log(l, calldepth, r)
}

func initLogging() {
	var (
		backend   logging.Backend
		logFormat logging.Formatter
		err       error
	)

	if logFile == "-" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		logFormat = logging.MustStringFormatter(logFormatDefault)

	} else if logFile == "systemd" {
		backend = logging.NewLogBackend(os.Stderr, "", 0)
		logFormat = logging.MustStringFormatter(logFormatPlain)

	} else if logFile == "syslog" {
		syslogPrefix := path.Base(os.Args[0])
		backend, err = logging.NewSyslogBackendPriority(syslogPrefix, syslog.LOG_LOCAL4)
		if err != nil {
			log.Fatalf("error opening channel to syslog %v", err)
		}
		logFormat = logging.MustStringFormatter(logFormatSyslog)

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

	{
		backend = &avoidDuplicatesBackend{backend: backend}
		formatted := logging.NewBackendFormatter(backend, logFormat)
		levelled := logging.AddModuleLevel(formatted)
		levelled.SetLevel(logging.Level(logLevel), "main")
		logging.SetBackend(levelled)
	}

}

// vim: foldmethod=syntax
