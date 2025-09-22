#!/bin/bash

# Simple HTTP server for testing the chat widget
echo "Starting HTTP server for chat widget testing..."
echo "Open http://localhost:3001 in your browser to test the widget"
echo ""
echo "Make sure your Hith backend is running on localhost:8080"
echo "Press Ctrl+C to stop the server"
echo ""

# Use Python's built-in HTTP server
if command -v python3 &> /dev/null; then
    python3 -m http.server 3001
elif command -v python &> /dev/null; then
    python -m SimpleHTTPServer 3001
else
    echo "Error: Python not found. Please install Python to run the test server."
    exit 1
fi
