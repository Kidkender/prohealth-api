#!/bin/sh
set -e

echo "Running seed..."
./seed

echo "Starting server..."
exec ./server
