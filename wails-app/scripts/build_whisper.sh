#!/usr/bin/env bash
set -e

# Setup directories
PROJECT_ROOT=$(cd $(dirname "$0")/.. && pwd)
BACKEND_DIR="$PROJECT_ROOT/backend"
LIB_DIR="$BACKEND_DIR/lib"
TMP_DIR=$(mktemp -d)

mkdir -p "$LIB_DIR"

echo "Cloning whisper.cpp..."
git clone https://github.com/ggml-org/whisper.cpp.git "$TMP_DIR/whisper.cpp"
cd "$TMP_DIR/whisper.cpp"
git checkout v1.9.4

echo "Building whisper.cpp..."
cd bindings/go
make whisper

echo "Copying libraries and headers..."
cp ../../build_go/src/libwhisper.a "$LIB_DIR/"
cp ../../build_go/ggml/src/libggml.a "$LIB_DIR/"
cp ../../build_go/ggml/src/libggml-base.a "$LIB_DIR/"
cp ../../build_go/ggml/src/libggml-cpu.a "$LIB_DIR/"

cp ../../include/whisper.h "$LIB_DIR/"
cp ../../ggml/include/ggml.h "$LIB_DIR/"
cp ../../ggml/include/ggml-cpu.h "$LIB_DIR/"
cp ../../ggml/include/ggml-alloc.h "$LIB_DIR/"
cp ../../ggml/include/ggml-backend.h "$LIB_DIR/"

echo "Cleaning up..."
rm -rf "$TMP_DIR"

echo "Done! Whisper libraries and headers are in $LIB_DIR"
