package main

import (
	"fsnotify"
	"log"
	"path"
)

/*
	Файл запишется в выходную очередь только в том случае, если
	1) Он был создан и затем
	2) он был закрыт
*/

func filterNewFiles(fileQueue chan string, watcher *fsnotify.Watcher) {
	defer watcher.Close()
	createdOrClosed := make(map[string]string)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Could not read event:", event, ok)
			} else {
				fn := path.Base(event.Name)
				log.Println("New event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("File was created:", fn)
					createdOrClosed[fn] = EMPTY_VALUE
				} else if event.Op&fsnotify.Close == fsnotify.Close {
					log.Println("File was closed:", event)
					if _, found := createdOrClosed[fn]; found {
						delete(createdOrClosed, fn)
						fileQueue <- fn
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
