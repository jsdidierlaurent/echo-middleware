#!/usr/bin/env bash
# This script require Docker / Docker-Compose

echo "Starting Redis/Memcached with Docker ..."
docker-compose up -d

echo "Running tests ..."
go clean -testcache `go list ./... | grep -v example`
go test -v `go list ./... | grep -v example`

echo "Killing Redis/Memcached ..."
docker-compose down
