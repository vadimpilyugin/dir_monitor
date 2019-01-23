package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const (
	SENT_LIST_FN   = "sent_list.txt"
	INIT_LIST_SIZE = 16
	EMPTY_VALUE    = ""
	PERM_ALL       = 0777
	MODE_APPEND    = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	MODE_CREATE    = os.O_CREATE | os.O_RDONLY
	ENDL           = "\n"
)

func EnqueueDir(dirPath string, fileQueue chan string, readyQueue chan string) {
	sentList := readSentList(dirPath)
	sentList[SENT_LIST_FN] = EMPTY_VALUE

	go func() {
		for _, fn := range dirList(dirPath) {
			if _, found := sentList[fn]; !found {
				fileQueue <- fn
			}
		}
	}()
	go writeReadyFiles(dirPath, readyQueue)
}

func readSentList(dirPath string) map[string]string {
	sentListFile, err := os.OpenFile(path.Join(dirPath, SENT_LIST_FN), MODE_CREATE, PERM_ALL)
	if err != nil {
		log.Fatal("Couldn't read sent_list", err)
	}
	defer sentListFile.Close()

	sentList := make(map[string]string)
	scanner := bufio.NewScanner(sentListFile)
	for scanner.Scan() {
		sentList[scanner.Text()] = EMPTY_VALUE
	}
	if err := scanner.Err(); err != nil {
		log.Println("Couldn't scan sent_list", err)
	}
	return sentList
}

func dirList(dirPath string) []string {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	list := make([]string, 0, INIT_LIST_SIZE)

	for _, file := range files {
		list = append(list, file.Name())
	}
	return list
}

func writeReadyFiles(dirPath string, readyQueue chan string) {
	sentListFile, err := os.OpenFile(path.Join(dirPath, SENT_LIST_FN), MODE_APPEND, PERM_ALL)
	if err != nil {
		log.Fatal(err)
	}
	defer sentListFile.Close()

	for {
		fn := <-readyQueue
		_, err := sentListFile.WriteString(fn + ENDL)
		if err != nil {
			log.Println("Error when writing to queueFile:", err)
		}
		sentListFile.Sync()
	}
}
