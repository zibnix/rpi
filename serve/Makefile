all:
		@go build -ldflags "-X main.version=$(shell git rev-parse --short --verify HEAD)" -o serve serve.go http.go sensor.go
race:
		@go build -race -ldflags "-X main.version=$(shell git rev-parse --short --verify HEAD)" -o serve serve.go http.go sensor.go
