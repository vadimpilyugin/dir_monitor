package main

import (
  "bytes"
  "log"
  "net/http"
  "time"
  "github.com/vadimpilyugin/http_over_at"
  "errors"
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

func sendFile(url string, dirPath, fn string) error {
	request, err := uploadRequest(url, dirPath, fn)
	if err != nil {
		log.Println("uploadRequest failed:", err)
		return err
	}
	client := &http.Client{
    Transport: http_over_at.Rqstr,
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
    if resp.StatusCode != http.StatusOK {
      return errors.New(http.StatusText(resp.StatusCode))
    }
		log.Println("--- ", resp.StatusCode)
		log.Println("--- ", resp.Header)
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
