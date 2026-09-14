#!/bin/bash
# Downloads nomic-embed-text-v1.5 INT8 quantized ONNX model + onnxruntime for embedding noise-gate
set -e
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODELS_DIR="$REPO_ROOT/wails-app/models/nomic-embed-text-v1.5"
LIB_DIR="$REPO_ROOT/wails-app"

echo "📥 Creating $MODELS_DIR"
mkdir -p "$MODELS_DIR"

echo "📥 Downloading model.onnx..."
curl -L -o "$MODELS_DIR/model.onnx" \
  "https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/onnx/model_quantized.onnx"

echo "📥 Downloading tokenizer.json..."
curl -L -o "$MODELS_DIR/tokenizer.json" \
  "https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/tokenizer.json"

echo "📥 Downloading onnxruntime v1.18.1..."
ONNX_ARCHIVE="onnxruntime-linux-x64-1.18.1.tgz"
curl -L -o "/tmp/$ONNX_ARCHIVE" \
  "https://github.com/microsoft/onnxruntime/releases/download/v1.18.1/$ONNX_ARCHIVE"
tar -xzf "/tmp/$ONNX_ARCHIVE" -C /tmp
cp "/tmp/onnxruntime-linux-x64-1.18.1/lib/libonnxruntime.so.1.18.1" "$LIB_DIR/libonnxruntime.so"
rm -rf "/tmp/$ONNX_ARCHIVE" "/tmp/onnxruntime-linux-x64-1.18.1"

echo "✅ Done! Run 'make dev' to start with embedding noise-gate enabled."
