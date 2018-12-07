package main

import (
	"log"
	"os"
)

func usage() {
	println("Usage: monitor [DIRECTORY] [POST_URL]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}
	log.Println("Hello, world!")
	dirToMonitor := os.Args[1]
	postUrl := os.Args[2]
	fileQueue := make(chan string)
	readyQueue := make(chan string)

	EnqueueDir(dirToMonitor, fileQueue, readyQueue)
	StartMonitor(dirToMonitor, fileQueue)
	SendFiles(dirToMonitor, postUrl, fileQueue, readyQueue)

	// done := make(chan bool)
	// go monitor(dirToMonitor, postUrl)
	// <-done
}
