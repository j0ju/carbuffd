// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import ()

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
	for _, r := range Cfg.Receivers {
		r.instance = CreateTcpCarbonReceiver(r.Url, metricsChannel)
		r.instance.Run()
	}

	go internalMetricsGenerator(metricsChannel)
	go metricsForwarder(Cfg.RemoteAddr, metricsChannel)

	<-doQuit
}

// vim: foldmethod=syntax
