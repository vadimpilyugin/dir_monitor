# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=monitor
BINARY_UNIX=$(BINARY_NAME)_unix

all: $(BINARY_NAME)
$(BINARY_NAME): main.go tls-client.go send_file.go file_manager.go config.go file_list.go monitor.go
	$(GOBUILD) -o $(BINARY_NAME) $^
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
deps:
	$(GOGET) github.com/vadimpilyugin/fsnotify
	$(GOGET) github.com/vadimpilyugin/http_over_at


# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

