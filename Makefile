# Makefile

.PHONY: build dev

build:
	cd wails-app && $(shell go env GOPATH)/bin/wails build -tags webkit2_41

dev:
	cd wails-app && $(shell go env GOPATH)/bin/wails dev -tags webkit2_41
