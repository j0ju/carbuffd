// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

const (
	SOCKET_TIMEOUT_DEFAULT = 90
	ERROR_WAIT_START_MSEC  = 100
)

var (
	metricsBufferSize uint64 = 10240
	laddr             string = ":2003"
	raddr             string = ""
	statsInterval     uint   = 60
	statsFmt          string = "carbon.carbuffd.%[1]s.%[2]s"
)

// vim: foldmethod=syntax
