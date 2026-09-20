# Makefile

.PHONY: build dev build-barnowl dev-barnowl build-presenter dev-presenter build-mac build-windows build-linux build-all

# BarnOwl AI (Interview Copilot) — Primary Product
build-barnowl:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41

dev-barnowl:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags webkit2_41

# Default aliases point to BarnOwl AI
build: build-barnowl
dev: dev-barnowl

# ==========================================
# Environment Setup & Database Migrations
# ==========================================
# How to run local vs Cloud/SaaS mode:
# 1. Local Mode: In `wails-app/frontend/.env`, leave PUBLIC_SUPABASE_URL empty or unset.
# 2. SaaS/Cloud Mode: In `wails-app/frontend/.env`, set PUBLIC_SUPABASE_URL and PUBLIC_SUPABASE_ANON_KEY to your Supabase instance (local or remote).
#
# To setup the local database (required for User Entitlements and OAuth):
supabase-setup-local:
	@echo "Pushing database schema to local Supabase instance..."
	sudo npx supabase db push
	@echo "Local database is now ready."

supabase-start:
	sudo npx supabase start

# Platform-specific builds
build-mac:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41 -platform darwin/universal

build-windows:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41 -platform windows/amd64

build-linux:
	cd wails-app && VITE_PRODUCT=interview C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags webkit2_41 -platform linux/amd64

build-all: build-mac build-windows build-linux

# StealthPresenter (Teleprompter Mode — Paused)
build-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails build -tags "webkit2_41 presenter"

dev-presenter:
	cd wails-app && VITE_PRODUCT=presenter C_INCLUDE_PATH=$(shell pwd)/wails-app/backend/lib LIBRARY_PATH=$(shell pwd)/wails-app/backend/lib CGO_CFLAGS="-I$(shell pwd)/wails-app/backend/lib" CGO_LDFLAGS="-L$(shell pwd)/wails-app/backend/lib -lwhisper -lstdc++ -lm" $(shell go env GOPATH)/bin/wails dev -tags "webkit2_41 presenter"

