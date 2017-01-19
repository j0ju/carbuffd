// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"fmt"
	"net"
	"time"
)

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

// vim: foldmethod=syntax
