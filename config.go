// LICENSE: GPLv2, see attached License
// Author: Joerg Jungermann

package main

import (
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
)

const (
	SOCKET_TIMEOUT_DEFAULT = 90
	ERROR_WAIT_START_MSEC  = 100
	// TODO: add Url scheme
	defaultConfigJsonFmt = `
  {
    "Receivers": [
      {
        "Url":  ""
      }
    ],
    "RemoteAddr":        "",
    "MetricsBufferSize": 10240,
    "LogDestination":    "-",
    "LogLevel":          %d,
    "StatsInterval":     60,
    "StatsFmt":          "%s"
  }
  `
)

var (
	defaultLogLevel = int(logging.NOTICE)
	defaultStatsFmt = "carbon.carbuffd.%[1]s.%[2]s"
	Cfg             *Config
)

type Config struct {
	LogDestination    string `json:"LogDestination",omitifempty`
	LogLevel          int    `json:"LogLevel",omitifempty`
	StatsInterval     uint   `json:"StatsInterval",omitifempty`
	StatsFmt          string `json:"StatsFmt",omitifempty`
	MetricsBufferSize uint64 `json:"MetricsBufferSize",omitifempty`

	Receivers  []*ConfigReceivers
	RemoteAddr string `json:"RemoteAddr",omitifempty`
}

type ConfigReceivers struct {
	Name string `json:"Name",omitifempty`
	Url  string `json:"Url",omitifempty`

	//  after unmarshalling call instance = CreateCarbonReceiver(Url)
	instance CarbonReceiver
}

func initConfig() {
	cfgJson := fmt.Sprintf(defaultConfigJsonFmt,
		defaultLogLevel,
		defaultStatsFmt)

	Cfg = new(Config)
	err := json.Unmarshal([]byte(cfgJson), Cfg)
	if err != nil {
		panic(err)
	}

}

// vim: foldmethod=syntax
