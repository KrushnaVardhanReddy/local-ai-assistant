# API Architecture Refactor

## Overview

Historically, BarnOwl AI exclusively used **Wails IPC bindings** (`window.go.main.App.*`) and the `EventsOn` WebSocket abstraction to communicate between the Svelte frontend and the Go backend.

To prepare for web-only deployments and decouple the frontend from the desktop application shell, a standard local **HTTP + WebSocket server** has been introduced in `wails-app/backend/api/`.

## The Hybrid Approach

The application currently supports a hybrid communication strategy:

1. **Local Desktop Mode (Wails)**: The desktop app binds a local HTTP server on a random port (sent via the `api_port` Wails event).
2. **Dev Mode / Web Mode**: The frontend reads the port from the `VITE_API_BASE_URL` environment variable (`http://localhost:8080`).

In `wails-app/frontend/src/lib/api.ts` and `ws.svelte.ts`, a `USE_REST` feature flag checks if the REST API path should be used instead of Wails IPC bindings. This allows the backend to be run headlessly or interacted with over HTTP while keeping backwards compatibility.

## API Endpoints

The new REST server (`backend/api/routes.go`) maps GET and POST requests directly to `AppInterface` methods (which `App` in `app.go` satisfies). All responses are wrapped in a generic JSON structure: `{"data": result}` or `{"error": err}`.

### Websocket Subsystem
Real-time events like `on_transcript`, `on_response_token`, `ptt-start`, and `buddy_hint` are now duplexed via a custom Gorilla websocket hub `backend/api/hub.go` accessible at `/ws`.

The Go function `a.emit(eventName, payload)` pushes standard Wails events *and* WebSocket payloads concurrently, meaning Web Clients and Wails Desktop Clients stay exactly in sync.