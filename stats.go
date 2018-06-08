// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type InternalStats struct {
	connectionCount        uint64
	currentConnectionCount uint64
	inConnectionTimeouts   uint64
	inConnectionErrors     uint64
	messageChannelLimit    uint64
	messageChannelSize     uint64
	messagesRelayed        uint64
	messagesDropped        uint64
	invalidMessages        uint64
	augmentedMessages      uint64
	uptimeSeconds          uint64
	outConnectionErrors    uint64
	numGoRoutines          uint64
	numGC                  uint64
	heapAlloc              uint64
}

var (
	stats InternalStats
)

func internalMetricsGenerator(ch chan string) {
	if statsFmt == "" { // if statsFmt is empty, do not emit metrics
		return
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	startEpoch := time.Now().Unix()
	hostname = strings.Replace(hostname, ".", "_", -1)
	tmpl := statsFmt + " %d %d"

	stats.messageChannelLimit = uint64(cap(ch))
	stats.numGoRoutines = uint64(runtime.NumGoroutine())
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	stats.heapAlloc = mem.HeapAlloc
	stats.numGC = uint64(mem.NumGC)
	for {
		time.Sleep(time.Duration(statsInterval) * time.Second)
		epoch := time.Now().Unix()

		stats.messageChannelSize = uint64(len(ch))
		stats.uptimeSeconds = uint64(epoch) - uint64(startEpoch)

		// metrics will be generated via reflection of internal stats struct
		reflectStats := reflect.ValueOf(&stats).Elem()
		for i := 0; i < reflectStats.NumField(); i++ {
			name := reflectStats.Type().Field(i).Name
			val := reflectStats.Field(i)
			metric := fmt.Sprintf(tmpl, hostname, name, val, epoch)

			if !(len(ch) < cap(ch)-1) {
				// dequeue old events to add newer events
				<-ch
				stats.messagesDropped++
				log.Warningf("dropped event, qlen %d ~ limit %d\n", len(ch), cap(ch))
			}
			ch <- metric
			log.Debugf("%s\n", metric)
		}
	}
}

// vim: foldmethod=syntax
