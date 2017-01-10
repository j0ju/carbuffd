package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	SOCKET_TIMEOUT_DEFAULT = 90
	ERROR_WAIT_START_MSEC  = 100
)

var (
	carbonClientHandler_count uint64 = 0
	metricsBufferSize         uint64 = 1048576
	laddr                     string = ":2003"
	raddr                     string = ""
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
		epoch = strconv.FormatInt(time.Now().Unix(), 10)
		//fmt.Printf("carbonMetricFilter: metric %s augmented with epoch %s\n", metric, epoch)
	}
	return fmt.Sprintf("%s %s %s", metric, value, epoch), true
}
func carbonClientHandler(c net.Conn, ch chan string) {
	fmt.Printf("carbonClientHandler[%d]: %s accepted\n", carbonClientHandler_count, c.RemoteAddr().String())
	carbonClientHandler_count++
	carbonClientHandler_metrics_ingress_count := 0

	// TODO: socket timeout setzen: max wert <- burst -> min wert
	// c.SetReadDeadline(!!! time.Duration(SOCKET_TIMEOUT_DEFAULT * time.Second))
	reader := bufio.NewReader(c)
	for {
		c.SetReadDeadline(time.Now().Add(SOCKET_TIMEOUT_DEFAULT * time.Second))
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Printf("carbonClientHandler[%d]: %s closed (%d lines received)\n", carbonClientHandler_count, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
			break
		} else if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				fmt.Printf("carbonClientHandler[%d]: %s timeout (%d lines received)\n", carbonClientHandler_count, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			} else if !netErr.Temporary() {
				fmt.Printf("carbonClientHandler[%d]: %s temporary (%d lines received)\n", carbonClientHandler_count, c.RemoteAddr().String(), carbonClientHandler_metrics_ingress_count)
				break
			}
		} else if err != nil {
			panic(err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// TODO metricsChannelReader in Go routine as a filter, it will put complete metrics into the channel
		metric, isCorrect := carbonMetricFilter(line)
		if !isCorrect {
			fmt.Printf("carbonClientHandler[%d]: non metric received '%s'\n", carbonClientHandler_count, line)
			continue
		}
		ch <- metric
		carbonClientHandler_metrics_ingress_count++
	}

	c.Close()
}
func carbonServer(laddr string, ch chan string) {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("carbonServer: listening on %s\n", laddr)
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go carbonClientHandler(c, ch)
	}
}
func connectToRemote(raddr string) (*net.Conn, error) {
	dialer := net.Dialer{
		Timeout: time.Duration(SOCKET_TIMEOUT_DEFAULT * time.Second),
	}
	d, e := dialer.Dial("tcp", raddr)
	return &d, e
}
func metricsChannelReader(raddr string, ch chan string) {
	var (
		c   *net.Conn
		err error
	)
	err_wait_msec := ERROR_WAIT_START_MSEC * time.Millisecond
	for m := range ch {
		// ensure connection
		if c == nil && raddr != "" {
			if c, err = connectToRemote(raddr); err != nil {
				c = nil
			}
		}
		// if we have a connection send it
		if c != nil {
			msg := []byte(m + "\n")
			_, err = (*c).Write(msg)
		}

		if netErr, ok := err.(net.Error); ok {
			if !netErr.Temporary() {
				fmt.Printf("metricsChannelReader: non temporary error, resetting socket\n")
				c = nil
			}
		}

		if err == nil {
			err_wait_msec = ERROR_WAIT_START_MSEC * time.Millisecond
			//fmt.Printf("%s\n", m)
		} else {
			time.Sleep(err_wait_msec)
			err_wait_msec *= 2
			ch <- m
			fmt.Printf("metricsChannelReader: %v, requeuing, %d queuelen, wait %v\n", err, len(ch), err_wait_msec)
		}
	}
}
func main() {

	flag.Parse()
	if len(flag.Args()) == 1 {
		laddr = flag.Args()[0]
	} else if len(flag.Args()) == 2 {
		laddr = flag.Args()[0]
		raddr = flag.Args()[1]
	} else {
		panic("to much/less command line arguments, need only two: laddr:lport raddr:rport")
	}

	metricsChannel := make(chan string, metricsBufferSize)
	go carbonServer(laddr, metricsChannel)
	metricsChannelReader(raddr, metricsChannel)
}

// vim: foldmethod=syntax
