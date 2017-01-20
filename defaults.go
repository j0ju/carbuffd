// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"github.com/op/go-logging"
)

const (
	SOCKET_TIMEOUT_DEFAULT = 90
	ERROR_WAIT_START_MSEC  = 100
)

var (
	laddr             string = "" // ":2003"
	raddr             string = ""
	metricsBufferSize uint64 = 10240
	statsInterval     uint   = 60
	statsFmt          string = "carbon.carbuffd.%[1]s.%[2]s"
	logLevel                 = int(logging.NOTICE)
	logFile           string = "-" // -, syslog, syslog:facility, /patch/to/file
)

// vim: foldmethod=syntax
