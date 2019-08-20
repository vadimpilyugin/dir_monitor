package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func getClientTlsConfig(serverCAFile string,
	certFile, keyFile string) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Println("Unable to load cert:", err)
		return nil, err
	}

	log.Println("Server CA:", serverCAFile)
	serverCACert, err := ioutil.ReadFile(serverCAFile)
	if err != nil {
		log.Println("Unable to open cert:", err)
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(serverCACert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func getSecureClient(serverCAFile string,
	certFile, keyFile string) (*http.Client, error) {

	tlsConfig, err := getClientTlsConfig(serverCAFile, certFile, keyFile)
	if err != nil {
		return nil, err
	}

	cl := &http.Client{
		Timeout: time.Second * 600,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   300 * time.Second,
				KeepAlive: 120 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       120 * time.Second,
			TLSHandshakeTimeout:   120 * time.Second,
			ExpectContinueTimeout: 120 * time.Second,
			TLSClientConfig:       tlsConfig,
		},
	}
	return cl, nil
}
