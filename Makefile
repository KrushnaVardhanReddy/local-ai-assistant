# Makefile

.PHONY: build dev

build:
	cd wails-app && wails build

dev:
	cd wails-app && wails dev
