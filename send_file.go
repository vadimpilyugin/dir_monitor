package main

import (
	"errors"
	"http_over_at"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
	// "context"
	runtime "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	strfmt "github.com/go-openapi/strfmt"
	apiclient "go-swagger-client/client"
	operations "go-swagger-client/client/operations"
	"os"
	"path"
)

const (
	PARAM_NAME       = "file"
	N_SECONDS        = 1
	DEFAULT_DURATION = 120
)

var (
	api *apiclient.Dirmon
)

func init() {
	transport := httptransport.New("localhost", "", nil)
	api = apiclient.New(transport, strfmt.Default)
}

func SendFiles(dirPath string, url string, fileQueue chan string, readyQueue chan string) {
	availableInterfaces()
	for {
		fn := <-fileQueue
		err := sendFile(url, dirPath, fn)
		if err != nil {
			log.Println("Failed to send file: ", err)
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
			client = &http.Client{
				CheckRedirect: noRedir,
			}
			up = true
			log.Println("--- Using interface: ", interfc)
			break
		}
	}
	if !up {
		client = &http.Client{
			Transport: http_over_at.Rqstr,
		}
		log.Println("--- Using USB interface")
	}
	return client
}

func sendFile(url string, dirPath, fn string) error {
	f, err := os.Open(path.Join(dirPath, fn))
	if err != nil {
		log.Fatal(err)
	}
	params := operations.NewUploadImagePostParamsWithHTTPClient(getClient())
	params.SetFile(runtime.NamedReader(fn, f))
	params.SetTimeout(DEFAULT_DURATION * time.Second)
	resp, err := api.Operations.UploadImagePost(params)
	if err != nil {
		return err
	}
	log.Println(resp.Payload)
	if !resp.Payload.Ok {
		return errors.New(resp.Payload.Descr)
	}
	return nil
}
