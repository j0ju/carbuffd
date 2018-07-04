// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
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

func metricsForwarder(raddr string, ch chan string) {
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
				stats.outConnectionErrors++
				log.Errorf("non temporary error, resetting socket\n")
				if c != nil {
					(*c).Close()
					c = nil
				}
			}
		}

		if err == nil {
			err_wait_msec = ERROR_WAIT_START_MSEC * time.Millisecond
			if raddr == "" {
				log.Debugf("%s\n", m)
			}
			stats.messagesRelayed++
		} else {
			time.Sleep(err_wait_msec)

			// limit exponential backoff to 30 minutes
			err_wait_msec *= 2
			if err_wait_msec > 1800*time.Second {
				err_wait_msec = 1800 * time.Second
			}

			// if channel is not full reinsert it
			if len(ch) < cap(ch) {
				ch <- m
				log.Warningf("%v, requeued, %d qlen, wait %v\n", err, len(ch), err_wait_msec)
			} else {
				stats.messagesDropped++
				log.Errorf("%v, not requeued, %d qlen ~ limit %d, wait %v\n", err, len(ch), cap(ch), err_wait_msec)
			}
		}
	}
}

// vim: foldmethod=syntax
