run_server:
	go run main.go

build:
	GIT_TERMINAL_PROMPT=1 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o bin/sample_socket_io main.go

clean:
	rm -r bin

