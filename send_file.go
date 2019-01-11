package main

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

const (
	PARAM_NAME = "file"
	N_SECONDS  = 1
)

func (fb FileBody) Close() error {
	err := fb.file.Close()
	return err
}

func SendFiles(dirPath string, url string, fileQueue chan string, readyQueue chan string) {
	for {
		fn := <-fileQueue
		go func() {
			err := sendFile(url, dirPath, fn)
			if err != nil {
				time.Sleep(N_SECONDS * time.Second) // if there is no connection, then wait
				fileQueue <- fn
			} else {
				readyQueue <- fn
			}
		}()
	}
}

func noRedir(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func sendFile(url string, dirPath, fn string) error {
	request, err := uploadRequest(url, dirPath, fn)
	if err != nil {
		log.Println("uploadRequest failed:", err)
		return err
	}
	client := &http.Client{
		CheckRedirect: noRedir,
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Println("client.Do failed:", err)
		return err
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Println("body.ReadFrom failed:", err)
			return err
		}
		resp.Body.Close()
		log.Println("--- ", resp.StatusCode)
		log.Println("--- ", resp.Header)
		if resp.StatusCode != http.StatusSeeOther {
			log.Println("Wrong status code")
		}
	}
	return nil
}

func uploadRequest(uri string, dirPath, fn string) (*http.Request, error) {
	fb, err := NewBody(dirPath, fn)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, fb.Reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", fb.ContentType)
	req.ContentLength = fb.Length
	return req, nil
}
