// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func carbonMetricFilter(l string) (string, bool) {
	s := strings.SplitN(l, " ", 2)
	if len(s) < 2 {
		return "", false
	}
	// TODO: error handling, what if this is not a metric?
	metric := s[0]
	l = strings.TrimSpace(s[1])

	s = strings.SplitN(l, " ", 2)
	value := s[0]

	epoch := ""
	if len(s) > 1 {
		epoch = s[1]
	}
	_, err := strconv.ParseUint(epoch, 10, 64)
	if err != nil {
		stats.augmentedMessages++
		epoch = strconv.FormatInt(time.Now().Unix(), 10)
		//fmt.Printf("carbonMetricFilter: metric %s augmented with epoch %s\n", metric, epoch)
	}
	return fmt.Sprintf("%s %s %s", metric, value, epoch), true
}
func carbonClientHandler(c net.Conn, ch chan string) {
	fmt.Printf("carbonClientHandler[%d]: %s accepted\n", stats.connectionCount, c.RemoteAddr().String())
	stats.connectionCount++
	stats.currentConnectionCount++
	carbonClientHandler_metrics_ingress_count := 0

	// TODO: socket timeout setzen: max wert <- burst -> min wert
	// c.SetReadDeadline(!!! time.Duration(SOCKET_TIMEOUT_DEFAULT * time.Second))
	reader := bufio.NewReader(c)
	for {
		c.SetReadDeadline(time.Now().Add(SOCKET_TIMEOUT_DEFAULT * time.Second))
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Printf("carbonClientHandler[%d]: %s closed (%d lines received)\n", stats.connectionCount, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
			break
		} else if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				fmt.Printf("carbonClientHandler[%d]: %s timeout (%d lines received)\n", stats.connectionCount, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			} else if !netErr.Temporary() {
				fmt.Printf("carbonClientHandler[%d]: %s temporary (%d lines received)\n", stats.connectionCount, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			}
		} else if err != nil {
			panic(err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		metric, isCorrect := carbonMetricFilter(line)
		if !isCorrect {
			fmt.Printf("carbonClientHandler[%d]: non metric received '%s' from %s\n", stats.connectionCount, line, c.RemoteAddr().String())
			stats.invalidMessages++
			continue
		}
		if !(uint64(len(ch)) < metricsBufferSize-1) {
			// dequeue old events to add newer events
			<-ch
			stats.messagesDropped++
			fmt.Printf("metricsChannelReader: dropped event, queuelen %d ~ limit %d\n", len(ch), metricsBufferSize)
		}
		ch <- metric
		carbonClientHandler_metrics_ingress_count++
	}

	c.Close()
	stats.currentConnectionCount--
}
func carbonServer(laddr string, ch chan string) {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	fmt.Printf("carbonServer: listening on %s\n", laddr)

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go carbonClientHandler(c, ch)
	}
}

// vim: foldmethod=syntax
