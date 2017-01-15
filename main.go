package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	SOCKET_TIMEOUT_DEFAULT = 90
	ERROR_WAIT_START_MSEC  = 100
)

type InternalStats struct {
	connectionCount        uint64
	currentConnectionCount uint64
	messagesRelayed        uint64
	messagesDropped        uint64
	invalidMessages        uint64
	augmentedMessages      uint64
}

var (
	metricsBufferSize uint64 = 1048576
	laddr             string = ":2003"
	raddr             string = ""
	statsInterval     uint   = 60
	statsFmt          string = "carbon.carbuffd.%[1]s.%[2]s"
	stats             InternalStats
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
				(*c).Close()
				c = nil
			}
		}

		if err == nil {
			err_wait_msec = ERROR_WAIT_START_MSEC * time.Millisecond
			if raddr == "" {
				fmt.Printf("%s\n", m)
			}
			stats.messagesRelayed++
		} else {
			time.Sleep(err_wait_msec)
			err_wait_msec *= 2
			// if channel is not full reinsert it
			if uint64(len(ch)) < metricsBufferSize-1 {
				ch <- m
				fmt.Printf("metricsChannelReader: %v, requeued, %d queuelen, wait %v\n", err, len(ch), err_wait_msec)
			} else {
				stats.messagesDropped++
				fmt.Printf("metricsChannelReader: %v, not requeued, %d queuelen ~ limit %d, wait %v\n", err, len(ch), metricsBufferSize, err_wait_msec)
			}
		}
	}
}
func internalMetricsGenerator(ch chan string) {
	if statsFmt == "" { // if statsFmt is empty, do not emit metrics
		return
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	hostname = strings.Replace(hostname, ".", "_", -1)
	tmpl := statsFmt + " %d %d"
	for {
		time.Sleep(time.Duration(statsInterval) * time.Second)
		epoch := time.Now().Unix()

		// metrics will be generated via reflection of internal stats struct
		reflectStats := reflect.ValueOf(&stats).Elem()
		for i := 0; i < reflectStats.NumField(); i++ {
			name := reflectStats.Type().Field(i).Name
			val := reflectStats.Field(i)
			metric := fmt.Sprintf(tmpl, hostname, name, val, epoch)

			if !(uint64(len(ch)) < metricsBufferSize-1) {
				// dequeue old events to add newer events
				<-ch
				stats.messagesDropped++
				fmt.Printf("internalMetricsGenerator: dropped event, queuelen %d ~ limit %d\n", len(ch), metricsBufferSize)
			}
			ch <- metric
			//fmt.Printf("internalMetricsGenerator: %s\n", metric)
		}
	}
}
func main() {

	flag.Uint64Var(&metricsBufferSize, "l", metricsBufferSize, "default queue len")
	flag.UintVar(&statsInterval, "i", statsInterval, "interval for internal metrics")
	flag.StringVar(&statsFmt, "p", statsFmt, "format for internal statistics %[1]s = HOSTNAME, %[2]s is the metric name, if empty no internal metrics will be generated")
	flag.Parse()

	if len(flag.Args()) == 1 { // test mode receive only
		laddr = flag.Args()[0]
	} else if len(flag.Args()) == 2 { // receive, augment, reley
		laddr = flag.Args()[0]
		raddr = flag.Args()[1]
	} else {
		panic("to much/less command line arguments, need only two: laddr:lport raddr:rport")
	}

	metricsChannel := make(chan string, metricsBufferSize)
	go carbonServer(laddr, metricsChannel)
	go internalMetricsGenerator(metricsChannel)
	metricsChannelReader(raddr, metricsChannel)
}

// vim: foldmethod=syntax
