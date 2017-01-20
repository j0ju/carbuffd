// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"path"
)

func parseCommandLine() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s <options> [listenIP:listenport] [remoteIP:port]\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Notes:\n")
		fmt.Fprintf(os.Stderr, "  If remoteIP:port is not given, it just consumes all metrics\n")
		fmt.Fprintf(os.Stderr, "  This is handy for debugging\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  Log level values\n")
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.DEBUG)
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.INFO)
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.NOTICE)
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.WARNING)
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.ERROR)
		fmt.Fprintf(os.Stderr, "    %[1]d %10[1]s \n", logging.CRITICAL)
		fmt.Fprintf(os.Stderr, " Only messages with lower or equal log level are logged.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	// defaults are set in default.go
	flag.Uint64Var(&metricsBufferSize, "size", metricsBufferSize, "default queue len")
	flag.UintVar(&statsInterval, "interval", statsInterval, "interval for internal metrics")
	flag.IntVar(&logLevel, "loglevel", logLevel, "log level")
	flag.StringVar(&logFile, "logfile", logFile, "log file, instead of logging to STDERR")
	flag.Parse()

	if len(flag.Args()) == 1 { // test mode receive only
		laddr = flag.Args()[0]
	} else if len(flag.Args()) == 2 { // receive, augment, reley
		laddr = flag.Args()[0]
		raddr = flag.Args()[1]
	} else {
		flag.Usage()
		os.Exit(1)
	}

}

// vim: foldmethod=syntax
