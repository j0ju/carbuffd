// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"github.com/op/go-logging"
	"os"
)

var (
	log       = logging.MustGetLogger("main")
	logFormat = logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05.000} %{level:8s} %{shortfunc}: %{message}`,
	)
)

func initLogging() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	formatted := logging.NewBackendFormatter(backend, logFormat)

	levelled := logging.AddModuleLevel(formatted)
	levelled.SetLevel(logging.Level(logLevel), "main")

	logging.SetBackend(levelled)
}

// vim: foldmethod=syntax
