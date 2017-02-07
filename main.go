// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import ()

func main() {
	parseCommandLine()
	initLogging()

	doQuit := make(chan bool, 1)

	metricsChannel := make(chan string, metricsBufferSize)
	carbonServer := CreateCarbonListener(laddr, metricsChannel)

	go carbonServer.Run()
	go internalMetricsGenerator(metricsChannel)
	go metricsForwarder(raddr, metricsChannel)

	<-doQuit
}

// vim: foldmethod=syntax
