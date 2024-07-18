#!/usr/bin/env make -f

PROJECTNAME := cfdnscli
SHELL := /bin/bash

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: build clean help

all: build

## build: Build the binary.
build: clean
	@go build -o $(PROJECTNAME) main.go

## clean: Cleanup.
clean:
	@-rm -f $(PROJECTNAME)

## help: Show this message.
help: Makefile
	@echo "Available targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
