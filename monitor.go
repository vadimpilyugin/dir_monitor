package main

import (
	"fsnotify"
	"log"
	"path"
)

/*
	Файл запишется в выходную очередь только в том случае, если
	1) Он был создан/записан и затем
	2) он был закрыт
*/

func filterNewFiles(inputQueue chan string, watcher *fsnotify.Watcher) {
	defer watcher.Close()
	createdOrClosed := make(map[string]string)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Could not read event:", event, ok)
			} else {
				fn := path.Base(event.Name)
				// log.Println("New event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("File was created:", fn)
					createdOrClosed[fn] = EMPTY_VALUE
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					// log.Println("File was written to:", fn)
					createdOrClosed[fn] = EMPTY_VALUE
				} else if event.Op&fsnotify.Close == fsnotify.Close {
					log.Println("File was closed:", event)
					if _, found := createdOrClosed[fn]; found {
						delete(createdOrClosed, fn)
						inputQueue <- fn
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Could not read error")
			}
			log.Println("Got error:", err)
		}
	}
}

func StartMonitor(dirPath string, inputQueue chan string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed:", err)
	}

	err = watcher.Add(dirPath)
	if err != nil {
		log.Fatal("watcher.Add failed: ", err)
	}

	go filterNewFiles(inputQueue, watcher)
}
