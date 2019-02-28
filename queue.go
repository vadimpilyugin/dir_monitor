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
	PERM_ALL       = 0644
	MODE_APPEND    = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	MODE_CREATE    = os.O_CREATE | os.O_RDONLY
	MODE_WRITE		 = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	ENDL           = "\n"
)

func EnqueueDir(dirPath string, deleteSent bool, fileQueue chan string, readyQueue chan string) {
	sentList := readSentList(dirPath)
	sentList[SENT_LIST_FN] = EMPTY_VALUE

	// put unsent files into the queue
	go func() {
		for _, fn := range dirList(dirPath) {
			if _, found := sentList[fn]; !found {
				fileQueue <- fn
			}
		}
	}()

	// make new list to remove old entries
	newSentList := make([]string, 0, len(sentList))
	for _, fn := range dirList(dirPath) {
		if _, found := sentList[fn]; found {
			newSentList = append(newSentList, fn)
		}
	}

	rewriteSentList(dirPath, newSentList)

	go writeReadyFiles(dirPath, deleteSent, readyQueue)
}

func readSentList(dirPath string) map[string]string {
	sentList := map[string]string{}

	sentListFile, err := os.OpenFile(path.Join(dirPath, SENT_LIST_FN), MODE_CREATE, PERM_ALL)
	if err != nil {
		log.Println("Couldn't read sent_list: ", err)
		return sentList
	}
	defer sentListFile.Close()

	scanner := bufio.NewScanner(sentListFile)
	for scanner.Scan() {
		sentList[scanner.Text()] = EMPTY_VALUE
	}
	if err := scanner.Err(); err != nil {
		log.Println("Couldn't scan sent_list: ", err)
	}
	return sentList
}

func rewriteSentList(dirPath string, newSentList []string) {
	sentListFile, err := os.OpenFile(path.Join(dirPath, SENT_LIST_FN), MODE_WRITE, PERM_ALL)
	if err != nil {
		log.Fatal("Error opening sent_list for rewriting: ", err)
	}
	for _, fn := range newSentList {
		_, err := sentListFile.WriteString(fn + ENDL)
		if err != nil {
			log.Println("Error when writing to sent_list: ", err)
		}
	}
	sentListFile.Sync()
}

func dirList(dirPath string) []string {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal("Could not list directory: ", err)
	}
	list := make([]string, 0, INIT_LIST_SIZE)

	for _, file := range files {
		list = append(list, file.Name())
	}
	return list
}

func writeReadyFiles(dirPath string, deleteSent bool, readyQueue chan string) {
	sentListFile, err := os.OpenFile(path.Join(dirPath, SENT_LIST_FN), MODE_APPEND, PERM_ALL)
	if err != nil {
		log.Fatal("Error opening sent_list for reading: ", err)
	}
	defer sentListFile.Close()

	for {
		fn := <-readyQueue
		if deleteSent {
			err := os.Remove(path.Join(dirPath, fn))
			if err != nil {
				log.Println("Could not remove sent file: ", err)
			}
		} else {
			_, err := sentListFile.WriteString(fn + ENDL)
			if err != nil {
				log.Println("Error when writing to sent_list: ", err)
			}
			sentListFile.Sync()
		}
	}
}
