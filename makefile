export PATH := $(GOPATH)/bin:$(PATH)

build:
	go build -v -o bin/benchyou bench/benchyou.go
