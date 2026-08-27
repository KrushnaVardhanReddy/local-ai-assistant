#!/bin/bash

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null
then
    echo "cloudflared could not be found."
    echo "Please install it by running the following command:"
    echo "curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb && sudo dpkg -i cloudflared.deb"
    exit 1
fi

echo "Starting FastAPI backend in the background..."
# Assuming we are running this from the repository root
cd backend
if [ -d "../venv" ]; then
    source ../venv/bin/activate
fi
uvicorn app:app --port 8000 &
BACKEND_PID=$!

echo "Starting Cloudflare tunnel..."
echo "========================================================"
echo "When the tunnel is up, copy the https://*.trycloudflare.com URL"
echo "and append '/helper' to the end."
echo "Send that full URL to your remote helper."
echo "========================================================"

cloudflared tunnel --url http://127.0.0.1:8000

# When cloudflared exits, kill the backend
kill $BACKEND_PID
