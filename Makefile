# Makefile

.PHONY: build dev

build:
	cd wails-app && C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41

dev:
	cd wails-app && C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags webkit2_41

build-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags "webkit2_41 presenter"

dev-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags "webkit2_41 presenter"
