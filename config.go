package main

import (
	"github.com/go-ini/ini"
	"log"
)

const (
	CONFIG_FN = "/home/pilyugin_v/go/src/dir_monitor/config.ini"
)

type Internal struct {
	Directory  string `ini:"dir_to_monitor"`
	DeleteSent bool   `ini:"delete_sent"`
	LogWriter  string `ini:"log_writer"`
	UseAT bool	`ini:"use_at"`
}

type QueueSettings struct {
	LimitQueueSize bool `ini:"limit_backlog"`
	MaxQueueSize int64 `ini:"backlog_max_size"`
	RemoveSmallFiles bool `ini:"remove_small_files"`
	MinRemoveSize int64 `ini:"min_remove_size"`
}

type Network struct {
	PostUrl string `ini:"post_url"`
	LimitBandwidth bool `ini:"limit_bandwidth"`
	WaitFor int `ini:"wait_for"`
	BufLen int `ini:"buflen"`
	SendTimeout int `ini:"send_timeout"`
	DialTimeout int `ini:"dial_timeout"`
	RetryWaitFor int `ini:"retry_wait_for"`
}

type Config struct {
	Internal `ini:"internal"`
	QueueSettings `ini:"queue"`
	Network  `ini:"network"`
}

func readConfig() *Config {
	config := new(Config)
	err := ini.MapTo(config, CONFIG_FN)
	if err != nil {
		log.Fatal("Could not map config file: ", err)
	}
	return config
}
