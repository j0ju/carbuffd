// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import ()

func main() {
	parseCommandLine()
	initLogging()

	metricsChannel := make(chan string, metricsBufferSize)
	go carbonServer(laddr, metricsChannel)
	go internalMetricsGenerator(metricsChannel)
	metricsChannelReader(raddr, metricsChannel)
}

// vim: foldmethod=syntax
