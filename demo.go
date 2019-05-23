package main

import (
  "log"
  "io/ioutil"
)

func main() {
  cfg := readConfig()

  client, err := getSecureClient(cfg.ServerCAFile, cfg.CertFile, cfg.KeyFile)
  if err != nil {
    log.Fatal("Could not get secure http client:", err)
  }

  resp, err := client.Get("https://myserver.fun:8080/")
  if err != nil {
    log.Fatal("Could not get data:", err)
  }

  data, _ := ioutil.ReadAll(resp.Body)
  defer resp.Body.Close()

  log.Printf("Data:\n--------------\n%s\n", data)
}