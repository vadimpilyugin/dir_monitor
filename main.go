package main

import (
	"flag"
	"log"
)

func setupCmd() (string, string, bool) {
	log.Println("Hello, world!")
	cfg := readConfig()
	dirToMonitor := flag.String("dir", cfg.Internal.Directory, "directory to monitor")
	postUrl := flag.String("url", cfg.Network.PostUrl, "where to send files")
	deleteSent := flag.Bool("cleanup", cfg.Internal.DeleteSent, "whether to delete sent files")
	flag.Parse()
	return *dirToMonitor, *postUrl, *deleteSent
}

func main() {
	dirToMonitor, postUrl, deleteSent := setupCmd()
	fileQueue := make(chan string)
	readyQueue := make(chan string)

	initApi(postUrl)

	EnqueueDir(dirToMonitor, deleteSent, fileQueue, readyQueue)
	StartMonitor(dirToMonitor, fileQueue)
	SendFiles(dirToMonitor, fileQueue, readyQueue)
}
