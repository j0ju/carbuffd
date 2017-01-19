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
		log.Infof("metric %s augmented with epoch %s\n", metric, epoch)
	}
	return fmt.Sprintf("%s %s %s", metric, value, epoch), true
}
func carbonClientHandler(c net.Conn, ch chan string) {
	log.Noticef("%s accepted (connection# %d)\n", c.RemoteAddr().String(), stats.connectionCount)
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
			log.Noticef("%s closed (%d lines received)\n", c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
			break
		} else if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				log.Errorf("%s timeout (%d lines received)\n", c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			} else if !netErr.Temporary() {
				log.Errorf("%s temporary (%d lines received)\n", c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
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
			log.Errorf("non metric received '%s' from %s\n", line, c.RemoteAddr().String())
			stats.invalidMessages++
			continue
		}
		log.Debugf("%s\n", line)
		if !(len(ch) < cap(ch)-1) {
			// dequeue old events to add newer events
			<-ch
			stats.messagesDropped++
			log.Warningf("dropped event, qlen %d ~ cap %d\n", len(ch), cap(ch))
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
	log.Noticef("listening on %s\n", laddr)

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go carbonClientHandler(c, ch)
	}
}

// vim: foldmethod=syntax
