FILE_LIST="main.go tls-client.go send_file.go file_manager.go config.go file_list.go monitor.go"
monitor_arm: $(FILE_LIST)
	GOOS=linux GOARCH=arm64 go build -o monitor_arm $^
monitor: $(FILE_LIST)
	GOOS=linux GOARCH=amd64 go build -o monitor $^
