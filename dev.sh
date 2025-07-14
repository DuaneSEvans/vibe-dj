#!/bin/bash

set -e

cleanup() {
    echo "Shutting down background services..."
    # Check if the OLLAMA_PID variable is set and refers to a running process
    if [ -n "$OLLAMA_PID" ] && kill -0 "$OLLAMA_PID" 2>/dev/null; then
        echo "Stopping Ollama server (PID: $OLLAMA_PID)..."
        kill "$OLLAMA_PID"
    fi
    echo "Cleanup complete."
}

# Register the cleanup function to be called on script exit (e.g., via Ctrl+C)
trap cleanup EXIT

echo "Starting Ollama server in the background..."
# Start ollama serve, redirecting its output to a log file
ollama serve > ollama.log 2>&1 &
OLLAMA_PID=$!

echo "Ollama server started with PID: $OLLAMA_PID. Waiting for it to initialize..."
sleep 5

echo "Ensuring 'llava' model is available and loading it into memory..."
# This command will pull the model if it's not present, then run a quick
# prompt to load it into VRAM so the first API call is fast.
ollama run llava "say hi" > /dev/null

echo "Starting Go web server..."
# Navigate to the server directory and run the main application
cd server && air