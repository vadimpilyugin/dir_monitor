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

func setupCmd() (string, string, string, bool, QueueSettings) {
	cfg := readConfig()
	dirToMonitor := flag.String("dir", cfg.Internal.Directory, "directory to monitor")
	postUrl := flag.String("url", cfg.Network.PostUrl, "where to send files")
	logWriter := flag.String("log", cfg.Internal.LogWriter,
		"where to write logs: either 'stdout' or 'syslog'")
	deleteSent := flag.Bool("cleanup", cfg.Internal.DeleteSent, "whether to delete sent files")
	flag.Parse()
	return *dirToMonitor, *postUrl, *logWriter, *deleteSent, cfg.QueueSettings
}

func setupLogger(logWriter string) {
	if logWriter == SYSLOG_WRITER {
		w, err := syslog.Dial(SYSLOG_PROTO, SYSLOG_ADDRESS,
			syslog.LOG_WARNING, TAG)
		if err != nil {
			log.Println("Couldn't connect to syslogd: ", err)
			return
		}
		log.SetOutput(w)
	}
}

func main() {
	dirToMonitor, postUrl, logWriter, deleteSent, qSet := setupCmd()
	setupLogger(logWriter)
	fileManager := InitFileManager(dirToMonitor, qSet)

	initApi(postUrl)

	EnqueueDir(dirToMonitor, fileManager.InputQueue)
	WriteReadyFiles(deleteSent, fileManager)
	StartMonitor(dirToMonitor, fileManager.InputQueue)
	SendFiles(fileManager)
}
