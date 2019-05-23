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
	"io"
)

const (
	DontWait = 0
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

func SendFiles(fileManager *FileManager, cfg *Config) {
	availableInterfaces()
	for fn := range fileManager.OutputQueue {
		err := sendFile(cfg, fn)
		if err != nil {
			log.Println("Failed to send file: ", err)
			if _, ok := err.(*os.PathError); ok {
				log.Println("Path error, removing file from queue")
				fileManager.RemoveQueue <- fn
				continue
			}
			// wait if there is no connection
			time.Sleep(time.Duration(cfg.RetryWaitFor) * time.Second) 
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
		log.Println("Could not open net.Interfaces: ", err)
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

var client *http.Client

func getClient(useAT bool, sendTimeout, dialTimeout int) *http.Client {

	client.Transport.(*http.Transport).DialContext = (&net.Dialer{
    Timeout:   time.Duration(dialTimeout) * time.Second,
    KeepAlive: 30 * time.Second,
    DualStack: true,
  }).DialContext

	up := false
	for _, interfc := range []string{"ppp0", "eth0", "enp3s0"} {
		if checkInterface(interfc) {
			client.CheckRedirect = noRedir
			up = true
			log.Println("--- Using interface: ", interfc)
			break
		}
	}
	if !up && useAT {
		// FIXME: SSL will stop working
		client.Transport = http_over_at.Rqstr
		log.Println("--- Using AT interface")
	} else if !up {
		log.Println("--- Not using AT interface and no interface is available!")
	}
	return client
}

func limitedWriter(f *os.File, pw *io.PipeWriter, buflen int, waitfor int) {

	buf := make([]byte, buflen)
	duration := time.Duration(waitfor) * time.Second
	var timer *time.Timer

OuterFor:
	for {
		nRead, err := f.Read(buf)
		if nRead == 0 && err == io.EOF {
			pw.Close()
			break OuterFor
		}
		if err != nil {
			log.Println("Error occured when reading from file:", err)
			pw.CloseWithError(err)
			break OuterFor
		}
		dataToWrite := buf[:nRead]
		for len(dataToWrite) > 0 {
			nWritten, err := pw.Write(dataToWrite)
			if err != nil {
				log.Println("Pipe reading end returned an error:", err)
				pw.CloseWithError(err)
				break OuterFor
			}
			if nWritten != len(dataToWrite) {
				log.Printf(
					"Warning: slow pipe reader! Available %d bytes, read %d bytes\n",
					len(dataToWrite), nWritten,
				)
			}
			dataToWrite = dataToWrite[nWritten:]
		}

		if timer == nil {
				timer = time.NewTimer(duration)
		} else {
			timer.Reset(duration)
		}

		<-timer.C
	}
}

func sendFile(cfg *Config, fn string) error {
	filePath := path.Join(cfg.Directory, fn)
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Could not open file to send:", err)
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Println("Could not get FileInfo: ", err)
		return err
	}
	log.Printf(
		"Trying to send file '%s' (size %d bytes), timeout=%d sec\n",
		info.Name(), info.Size(), cfg.SendTimeout,
	)

	pr, pw := io.Pipe()
	if cfg.LimitBandwidth {
		go limitedWriter(f, pw, cfg.BufLen, cfg.WaitFor)
	} else {
		go limitedWriter(f, pw, cfg.BufLen, DontWait)
	}

	params := operations.NewUploadImagePostParamsWithHTTPClient(
		getClient(cfg.UseAT, cfg.SendTimeout, cfg.DialTimeout),
	)
	params.SetFile(runtime.NamedReader(fn, pr))
	params.SetTimeout(time.Duration(cfg.SendTimeout) * time.Second)
	
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
