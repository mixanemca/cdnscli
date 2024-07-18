#!/usr/bin/env make -f

PROJECTNAME := cfdnscli
BUILD := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --abbrev=0 --tags)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-s -w -X 'github.com/version-go/ldflags.buildVersion=$(VERSION)' -X 'github.com/version-go/ldflags.buildHash=$(BUILD)'"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: build clean help

all: build

## build: Build the binary.
build: clean
	@go build $(LDFLAGS) -o $(PROJECTNAME)

## clean: Cleanup.
clean:
	@-rm -f $(PROJECTNAME)

## help: Show this message.
help: Makefile
	@echo "Available targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
