package main

import (
  "github.com/go-ini/ini"
  "log"
)

const (
  CONFIG_FN = "config.ini"
)

type Internal struct {
  Directory   string  `ini:"dir_to_monitor"`
  DeleteSent  bool    `ini:"delete_sent"`
}

type Network struct {
  PostUrl string `ini:"post_url"`
}

type Config struct {
  Internal `ini:"internal"`
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
