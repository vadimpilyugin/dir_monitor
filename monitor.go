package main

import (
	"github.com/vadimpilyugin/fsnotify"
	"log"
	"path"
)

func filterNewFiles(fileQueue chan string, watcher *fsnotify.Watcher) {
	defer watcher.Close()
	createClose := make(map[string]string)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Fatal("Could not read event:", event, ok)
			}
			fn := path.Base(event.Name)
			log.Println("Some event:", event)
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("Created file:", fn)
				createClose[fn] = EMPTY_VALUE
			} else if event.Op&fsnotify.Close == fsnotify.Close {
				log.Println("Closed file:", event)
				if _, found := createClose[fn]; found {
					delete(createClose, fn)
					fileQueue <- fn
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Fatal("Could not read error")
			}
			log.Fatal("Got error:", err)
		}
	}
}

func StartMonitor(path string, fileQueue chan string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed:", err)
	}

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	go filterNewFiles(fileQueue, watcher)
}
