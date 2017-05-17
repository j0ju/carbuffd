// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"fmt"
)

var (
	doQuit chan bool
)

func main() {
	initConfig()
	parseCommandLine()

	initLogging()
	initSignalHandling()

	doQuit = make(chan bool, 1)
	metricsChannel := make(chan string, Cfg.MetricsBufferSize)

	// start receivers
	for i, r := range Cfg.Receivers {
		fmt.Printf("%d: %v\n", i, r)
		r.instance = CreateTcpCarbonReceiver(r.Url, metricsChannel)
		go r.instance.Run()
	}

	go internalMetricsGenerator(metricsChannel)
	go metricsForwarder(Cfg.RemoteAddr, metricsChannel)

	<-doQuit
}

// vim: foldmethod=syntax
