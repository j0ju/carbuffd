// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import ()

func main() {
	parseCommandLine()
	initLogging()

	doQuit := make(chan bool, 1)

	metricsChannel := make(chan string, metricsBufferSize)
	go carbonServer(laddr, metricsChannel)
	go internalMetricsGenerator(metricsChannel)
	go metricsChannelReader(raddr, metricsChannel)

	<-doQuit
}

// vim: foldmethod=syntax
