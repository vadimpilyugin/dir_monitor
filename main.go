package main

import (
	"flag"
	"log"
	"log/syslog"
)

const (
	SYSLOG_WRITER  = "syslog"
	SYSLOG_ADDRESS = "localhost:514"
	SYSLOG_PROTO   = "udp"
	STDOUT_WRITER  = "stdout"
	TAG            = "dir_monitor"
)

func setupCmd() (*Config, string, string, string, bool) {
	cfg := readConfig()
	dirToMonitor := flag.String("dir", cfg.Internal.Directory, "directory to monitor")
	postUrl := flag.String("url", cfg.Network.PostUrl, "where to send files")
	logWriter := flag.String("log", cfg.Internal.LogWriter,
		"where to write logs: either 'stdout' or 'syslog'")
	deleteSent := flag.Bool("cleanup", cfg.Internal.DeleteSent, "whether to delete sent files")
	flag.Parse()
	return cfg, *dirToMonitor, *postUrl, *logWriter, *deleteSent
}

func setupLogger(logWriter string) {
	if logWriter == SYSLOG_WRITER {
		w, err := syslog.New(syslog.LOG_NOTICE, TAG)
		if err != nil {
			log.Println("Couldn't connect to syslogd: ", err)
			return
		}
		log.SetOutput(w)
	}
}

func main() {
	cfg, dirToMonitor, postUrl, logWriter, deleteSent := setupCmd()
	setupLogger(logWriter)
	log.Printf("cfg: %#v\n", cfg)

	var err error
	client, err = getSecureClient(cfg.ServerCAFile, cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal("Could not get secure http client:", err)
	}
	// client = &http.Client{}

	fileManager := InitFileManager(dirToMonitor, cfg.QueueSettings)

	initApi(postUrl)

	EnqueueDir(dirToMonitor, fileManager.InputQueue)
	WriteReadyFiles(deleteSent, fileManager)
	StartMonitor(dirToMonitor, fileManager.InputQueue)
	SendFiles(fileManager, cfg)
}
