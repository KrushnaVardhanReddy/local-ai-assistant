# Makefile

.PHONY: build dev build-barnowl dev-barnowl build-presenter dev-presenter

# BarnOwl AI (Interview Copilot) — Primary Product
build-barnowl:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41

dev-barnowl:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags webkit2_41

# Default aliases point to BarnOwl AI
build: build-barnowl
dev: dev-barnowl

# StealthPresenter (Teleprompter Mode — Paused)
build-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags "webkit2_41 presenter"

dev-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags "webkit2_41 presenter"

