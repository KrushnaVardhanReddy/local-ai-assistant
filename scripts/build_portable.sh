#!/usr/bin/env bash
# Build a portable Windows release: AudioService-portable-win.zip
# Run from repo root. Requires: pyinstaller, cargo, tauri-cli, zip

set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist/portable"
VERSION=$(grep '"version"' "$ROOT/frontend/src-tauri/tauri.conf.json" | head -1 | grep -oP '[\d.]+')

echo "═══════════════════════════════════════════"
echo " AudioService Portable Build v$VERSION"
echo "═══════════════════════════════════════════"

# Clean
rm -rf "$DIST"
mkdir -p "$DIST"

# 1. Build Python backend with PyInstaller
echo "[1/3] Building backend EXE..."
cd "$ROOT/backend"
../.venv/bin/pyinstaller \
    --onefile \
    --name AudioService-backend \
    --distpath "$DIST" \
    --workpath /tmp/pyinstaller_build \
    --specpath /tmp \
    --noconfirm \
    app.py
echo "      ✅ Backend: $DIST/AudioService-backend.exe"

# 2. Build Tauri portable app
echo "[2/3] Building Tauri portable app..."
cd "$ROOT/frontend"
npm run tauri build -- --no-bundle 2>/dev/null || \
    npm run tauri build
TAURI_APP="$ROOT/frontend/src-tauri/target/release/bundle/app"
if [ -d "$TAURI_APP" ]; then
    cp -r "$TAURI_APP"/. "$DIST/"
else
    # Fallback: copy just the EXE from release
    cp "$ROOT/frontend/src-tauri/target/release/AudioService.exe" "$DIST/"
fi
echo "      ✅ Frontend: $DIST/"

# 3. Copy .env.local template (without secrets)
echo "[3/3] Packaging..."
cp "$ROOT/.env.local.example" "$DIST/.env.local" 2>/dev/null || \
    grep -v "API_KEY\|SECRET\|KEY=" "$ROOT/.env.local" > "$DIST/.env.local"

# Zip everything
cd "$ROOT/dist"
zip -r "AudioService-portable-win-v${VERSION}.zip" portable/
echo ""
echo "═══════════════════════════════════════════"
echo " ✅ Release: dist/AudioService-portable-win-v${VERSION}.zip"
echo "═══════════════════════════════════════════"
