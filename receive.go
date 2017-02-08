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

type CarbonListener struct {
	laddr          *net.TCPAddr
	socket         *net.TCPListener
	messageChannel chan string
	running        bool
}

func CreateCarbonListener(laddr string, ch chan string) *CarbonListener {
	var err error
	l := new(CarbonListener)
	l.laddr, err = net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	l.messageChannel = ch
	return l
}

func (l *CarbonListener) Stop() {
	err := l.socket.SetDeadline(time.Now().Add(1 * time.Millisecond))
	l.running = false
	if err != nil {
		log.Critical(err)
		panic(err)
	}
}

func (l *CarbonListener) Run() {
	var err error
	l.socket, err = net.ListenTCP("tcp", l.laddr)
	if err != nil {
		log.Critical(err)
		panic(err)
	}
	defer l.socket.Close()

	l.running = true
	log.Noticef("listening on %s\n", laddr)
	for l.running {
		c, err := l.socket.Accept()
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				if l.running {
					log.Errorf("listener on %s: %s\n", l.laddr, err)
				}
				continue
			}
		} else if err != nil {
			log.Critical(err)
			break
		}
		go l.clientHandler(c)
	}
}

// TODO: remove global stats dependency
func (l *CarbonListener) clientHandler(c net.Conn) {
	log.Noticef("%s accepted (connection# %d)\n", c.RemoteAddr().String(), stats.connectionCount)
	stats.connectionCount++
	stats.currentConnectionCount++
	carbonClientHandler_metrics_ingress_count := 0

	reader := bufio.NewReader(c)
	for { // endless loop, break on close or errors
		// TODO: socket timeout setzen: max wert <- burst -> min wert
		c.SetReadDeadline(time.Now().Add(SOCKET_TIMEOUT_DEFAULT * time.Second))
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Noticef("%s closed (%d lines received)\n", c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
			break
		} else if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				stats.inConnectionTimeouts++
				log.Errorf("%s timeout (%d lines received)\n", c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			} else if !netErr.Temporary() {
				stats.inConnectionErrors++
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

		metric, isMetric := l.metricFilter(line)
		if !isMetric {
			log.Errorf("non metric received '%s' from %s\n", line, c.RemoteAddr().String())
			stats.invalidMessages++
			continue
		}
		log.Debugf("%s\n", line)
		if !(len(l.messageChannel) < cap(l.messageChannel)-1) {
			// dequeue old events to add newer events
			<-l.messageChannel
			stats.messagesDropped++
			log.Warningf("dropped event, qlen %d ~ cap %d\n", len(l.messageChannel), cap(l.messageChannel))
		}
		l.messageChannel <- metric
		carbonClientHandler_metrics_ingress_count++
	}

	c.Close()
	stats.currentConnectionCount--
}

func (l *CarbonListener) metricFilter(m string) (string, bool) {
	s := strings.SplitN(m, " ", 2)
	if len(s) < 2 {
		return "", false
	}
	// TODO: error handling, what if this is not a metric?
	metric := s[0]
	m = strings.TrimSpace(s[1])

	s = strings.SplitN(m, " ", 2)
	value := s[0]

	epoch := ""
	if len(s) > 1 {
		epoch = s[1]
	}
	_, err := strconv.ParseUint(epoch, 10, 64)
	if err != nil {
		stats.augmentedMessages++
		epoch = strconv.FormatInt(time.Now().Unix(), 10)
		log.Debugf("metric %s augmented with epoch %s\n", metric, epoch)
	}
	return fmt.Sprintf("%s %s %s", metric, value, epoch), true
}

// vim: foldmethod=syntax
