#!/usr/bin/env make -f

PROJECTNAME := cdnscli
BUILD := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --abbrev=0 --tags)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-s -w -X 'github.com/version-go/ldflags.buildVersion=$(VERSION)' -X 'github.com/version-go/ldflags.buildHash=$(BUILD)'"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: build clean help

all: build

## lint: Run linting.
lint:
	@golangci-lint run ./...
	@staticcheck ./...
	@errcheck ./...
	@revive -config .revive.toml -formatter friendly ./...

## test: Run tests.
test: lint
	@go test -v -coverprofile=coverage.out github.com/mixanemca/$(PROJECTNAME)/internal/providers
	@go tool cover -func=coverage.out

## build: Build the binary.
build: clean test
	@go build $(LDFLAGS) -o $(PROJECTNAME)

## clean: Cleanup.
clean:
	@-rm -f $(PROJECTNAME)

## help: Show this message.
help: Makefile
	@echo "Available targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
