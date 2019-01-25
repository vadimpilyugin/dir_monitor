package main

import (
  "bytes"
  "log"
  "net/http"
  "time"
  "github.com/vadimpilyugin/http_over_at"
  "errors"
  "net"
  "strings"
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
	availableInterfaces()
	for {
		fn := <-fileQueue
		// go func() {
			err := sendFile(url, dirPath, fn)
			if err != nil {
				log.Println("Failed to send file", err)
				time.Sleep(N_SECONDS * time.Second) // if there is no connection, then wait
				go func() {
					fileQueue <- fn
				}()
			} else {
				readyQueue <- fn
			}
		// }()
	}
}

func noRedir(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func availableInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Available network interfaces on this machine")
	for _, i := range interfaces {
		log.Println("--- Interface name: ", i.Name)
	}
}

func checkInterface(ifName string) bool {
	byNameInterface, err := net.InterfaceByName(ifName)
	if err != nil {
		log.Println("["+ifName+"]", err)
		return false
	}
	if strings.Contains(byNameInterface.Flags.String(), "up") {
		return true
	} else {
		return false
	}
}

func sendFile(url string, dirPath, fn string) error {
	request, err := uploadRequest(url, dirPath, fn)
	if err != nil {
		log.Println("uploadRequest failed:", err)
		return err
	}
	var client *http.Client
	up := false
	for _, interfc := range []string{"ppp0", "eth0", "enp3s0"} {
		if checkInterface(interfc) {
			client = &http.Client {
				CheckRedirect: noRedir,
			}
			up = true
			log.Println("--- Using interface: ", interfc)
			break
		}
	}
	if !up {
		client = &http.Client {
			Transport: http_over_at.Rqstr,
		}
		log.Println("--- Using USB interface")
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
			return errors.New("Wrong status code")
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
