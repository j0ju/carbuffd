// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"flag"
)

func main() {

	flag.Uint64Var(&metricsBufferSize, "l", metricsBufferSize, "default queue len")
	flag.UintVar(&statsInterval, "i", statsInterval, "interval for internal metrics")
	flag.StringVar(&statsFmt, "p", statsFmt, "format for internal statistics %[1]s = HOSTNAME, %[2]s is the metric name, if empty no internal metrics will be generated")
	flag.Parse()

	if len(flag.Args()) == 1 { // test mode receive only
		laddr = flag.Args()[0]
	} else if len(flag.Args()) == 2 { // receive, augment, reley
		laddr = flag.Args()[0]
		raddr = flag.Args()[1]
	} else {
		panic("to much/less command line arguments, need only two: laddr:lport raddr:rport")
	}

	metricsChannel := make(chan string, metricsBufferSize)
	go carbonServer(laddr, metricsChannel)
	go internalMetricsGenerator(metricsChannel)
	metricsChannelReader(raddr, metricsChannel)
}

// vim: foldmethod=syntax
