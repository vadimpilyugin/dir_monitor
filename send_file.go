package main

import (
	"errors"
	runtime "github.com/go-openapi/runtime"
	strfmt "github.com/go-openapi/strfmt"
	apiclient "go-swagger-client/client"
	operations "go-swagger-client/client/operations"
	"http_over_at"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	PARAM_NAME     = "file"
	N_SECONDS      = 1
	DefaultTimeout = 600
)

var (
	api *apiclient.Dirmon
)

func initApi(postUrl string) {
	cfg := apiclient.DefaultTransportConfig()
	cfg.Host = postUrl
	api = apiclient.NewHTTPClientWithConfig(
		strfmt.Default,
		cfg,
	)
}

func SendFiles(fileManager *FileManager, useAT bool) {
	availableInterfaces()
	for fn := range fileManager.OutputQueue {
		err := sendFile(fileManager.dirPath, fn, useAT)
		if err != nil {
			log.Println("Failed to send file: ", err)
			if _, ok := err.(*os.PathError); ok {
				log.Println("Path error, removing file from queue")
				fileManager.RemoveQueue <- fn
				continue
			}
			time.Sleep(N_SECONDS * time.Second) // if there is no connection, then wait
			fileManager.PutBackCh <- fn
		} else {
			fileManager.ReadyQueue <- fn
		}
	}
}

func noRedir(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func availableInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("Could not open net.Interfaces: ", err)
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

func getClient(useAT bool) *http.Client {
	var client *http.Client
	up := false
	for _, interfc := range []string{"ppp0", "eth0", "enp3s0"} {
		if checkInterface(interfc) {
			client = &http.Client{
				CheckRedirect: noRedir,
			}
			up = true
			log.Println("--- Using interface: ", interfc)
			break
		}
	}
	if !up && useAT {
		client = &http.Client{
			Transport: http_over_at.Rqstr,
		}
		log.Println("--- Using USB interface")
	} else {
		log.Println("--- Not using AT interface and no interface is available!")
	}
	return client
}

func sendFile(dirPath, fn string, useAT bool) error {
	f, err := os.Open(path.Join(dirPath, fn))
	if err != nil {
		log.Println("Could not open file to send: ", err)
		return err
	}
	params := operations.NewUploadImagePostParamsWithHTTPClient(getClient(useAT))
	params.SetFile(runtime.NamedReader(fn, f))
	params.SetTimeout(DefaultTimeout * time.Second)
	info, err := f.Stat()
	if err != nil {
		log.Println("Could not get FileInfo: ", err)
		return err
	}
	log.Printf(
		"Trying to send file '%s' (size %d bytes), timeout=%d sec\n",
		info.Name(), info.Size(), DefaultTimeout,
	)
	resp, err := api.Operations.UploadImagePost(params)
	if err != nil {
		log.Println("UploadImagePost failed: ", err)
		return err
	}
	if !resp.Payload.Ok {
		log.Println(resp.Payload)
		log.Println("JSON payload OK is false")
		return errors.New(resp.Payload.Descr)
	}
	log.Println("Sent file:", fn)
	return nil
}
