#!/bin/sh
set -e

echo "Running migrations..."
./migrate -path ./migrations -database "$DATABASE_URL" up

echo "Running seed..."
./seed

echo "Starting server..."
exec ./server
