// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	c chan os.Signal
)

//TODO:
func handleSignal() {
	for s := range c {
		log.Notice("%v\n", s)
		fmt.Printf("%v\n", s)
		switch s {
		default:
		}
	}
}

func initSignalHandling() {
	c = make(chan os.Signal, 1)
	go handleSignal()

	// signals to catch
	//signal.Notify(c, os.Interrupt)
	//signal.Notify(c, syscall.SIGHUP)
	//signal.Notify(c, syscall.SIGQUIT)
	//signal.Notify(c, syscall.SIGTERM)
	//signal.Notify(c, syscall.SIGUSR1)
	//signal.Notify(c, syscall.SIGUSR2)
}
