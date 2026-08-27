#!/bin/bash
set -e

REPO_ROOT=$(dirname "$(dirname "$(readlink -f "$0")")")
cd "$REPO_ROOT" || exit 1

# ──────────────────────────────────────────────────────────────────────────────
# build.sh — Production build for Local AI Assistant
#
# Usage:
#   bash scripts/build.sh              # Build both backend and frontend
#   bash scripts/build.sh --backend    # Python backend only (PyInstaller)
#   bash scripts/build.sh --frontend   # Tauri + Svelte frontend only
# ──────────────────────────────────────────────────────────────────────────────

# Detect OS suffix for binary naming
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS_NAME" in
  *mingw*|*msys*|*cygwin*) SUFFIX="-windows.exe" ;;
  darwin)                   SUFFIX="-macos" ;;
  *)                        SUFFIX="-linux" ;;
esac

BUILD_BACKEND=true
BUILD_FRONTEND=true

if [ "$1" == "--backend" ]; then
    BUILD_FRONTEND=false
elif [ "$1" == "--frontend" ]; then
    BUILD_BACKEND=false
fi

# ── Backend: package Python app as a standalone binary ───────────────────────
if [ "$BUILD_BACKEND" == "true" ]; then
    echo "🐍 Building Python backend..."
    cd "$REPO_ROOT/backend"

    if ! command -v pyinstaller &>/dev/null; then
        echo "⚠️  pyinstaller not found. Installing..."
        pip install pyinstaller
    fi

    pyinstaller \
        --name "local-ai-backend${SUFFIX}" \
        --onefile \
        --hidden-import fastapi \
        --hidden-import uvicorn \
        --hidden-import faster_whisper \
        --hidden-import websockets \
        --hidden-import sounddevice \
        --hidden-import chromadb \
        --hidden-import sentence_transformers \
        app.py

    echo "✅ Backend binary: backend/dist/local-ai-backend${SUFFIX}"
    cd "$REPO_ROOT"
fi

# ── Frontend: Tauri production build ─────────────────────────────────────────
if [ "$BUILD_FRONTEND" == "true" ]; then
    echo "🖥️  Building Svelte 5 + Tauri 2.0 frontend..."
    cd "$REPO_ROOT/frontend"

    if [ ! -d "node_modules" ]; then
        echo "📦 Installing Node dependencies..."
        npm install
    fi

    # Required for linuxdeploy on some CI systems
    export APPIMAGE_EXTRACT_AND_RUN=1

    npm run tauri build

    echo "✅ Tauri bundle: frontend/src-tauri/target/release/bundle/"
    cd "$REPO_ROOT"
fi

echo ""
echo "🎉 Build complete!"
echo "   Backend:  backend/dist/local-ai-backend${SUFFIX}"
echo "   Frontend: frontend/src-tauri/target/release/"
