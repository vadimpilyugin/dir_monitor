package main

import (
  "net/http"
  "bytes"
  "log"
  "io"
  "mime/multipart"
  "os"
  "path/filepath"
  "path"
  "time"
)

const (
  PARAM_NAME = "file"
  N_SECONDS = 1
)

func SendFiles(dirPath string, url string, fileQueue chan string, readyQueue chan string) {
  for {
    fn := <-fileQueue
    go func () {
      err := sendFile(url, path.Join(dirPath, fn))
      if err != nil {
        time.Sleep(N_SECONDS * time.Second) // if there is no connection, then wait
        fileQueue <- fn
      } else {
        readyQueue <- fn
      }
    }()
  }
}

func sendFile (url string, name string) error {
  request, err := uploadRequest(url, name)
  if err != nil {
    log.Println("uploadRequest failed:", err)
    return err
  }
  client := &http.Client{}
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
  }
  return nil
}

func uploadRequest(uri string, full_path string) (*http.Request, error) {
  file, err := os.Open(full_path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)
  
  part, err := writer.CreateFormFile(PARAM_NAME, filepath.Base(full_path))
  if err != nil {
    return nil, err
  }

  _, err = io.Copy(part, file)
  if err != nil {
    return nil, err
  }

  err = writer.Close()
  if err != nil {
    return nil, err
  }

  req, err := http.NewRequest("POST", uri, body)
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", writer.FormDataContentType())
  return req, nil
}