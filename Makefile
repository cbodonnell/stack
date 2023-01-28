include .env

VERSION ?= $(shell git rev-parse --short HEAD)
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

build:
	go build \
	-ldflags="-X 'github.com/cbodonnell/stack/cmd/commands.version=${VERSION}'" \
	-o ./bin/stack ./cmd/main.go

clean:
	rm -rf ./bin
