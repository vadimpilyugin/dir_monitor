package main

import (
  "log"
  "net/http"
  "time"
  "http_over_at"
  "errors"
  "net"
  "strings"
  sw "dirmon_client"
  "os"
  "path"
)

const (
	PARAM_NAME = "file"
	N_SECONDS  = 1
)

var (
	cfg *sw.Configuration
)

func init() {
	cfg = sw.NewConfiguration()
}

func SendFiles(dirPath string, url string, fileQueue chan string, readyQueue chan string) {
	availableInterfaces()
	for {
		fn := <-fileQueue
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

func getClient() *http.Client {
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
	return client
}

func sendFile(url string, dirPath, fn string) error {
	
	cfg.HTTPClient = getClient()
	apiClient := sw.NewAPIClient(cfg)
	f, err := os.Open(path.Join(dirPath, fn))
	if err != nil {
		log.Fatal(err)
	}
	payload, resp, err := apiClient.DefaultApi.UploadImagePost(nil, map[string]interface{}{
    "image" : f,
  })
	if err != nil {
		log.Println("UploadImagePost failed:", err)
		return err
	} else {
		log.Println(payload, resp)
		if resp.StatusCode != http.StatusOK {
			log.Println("Wrong status code")
			return errors.New("Wrong status code")
		}
	}
	return nil
}
