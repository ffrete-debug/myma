#!/bin/sh
set -e

# Read config from environment with defaults
API_PORT="${SERVER_PORT:-:8080}"
FRONTEND_PORT="${PORT:-3000}"
DB_PATH="${DB_PATH:-/data/ark-commander.db}"

export DB_PATH
export SERVER_PORT="$API_PORT"
export PORT="$FRONTEND_PORT"

# Ensure data directory exists
mkdir -p /data

# Start Go API in background
echo "Starting ARK Commander API on $API_PORT..."
/app/bin/ark-commander &\
API_PID=$!

# Wait a moment for API to be ready
sleep 2

# Start Next.js frontend
echo "Starting Next.js frontend on port $FRONTEND_PORT..."
cd /app/ui
exec npm run start
