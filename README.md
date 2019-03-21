# Echo Middleware
Inspired by https://github.com/hb-go/echo-web

> Requires

[![Slack](https://img.shields.io/badge/Echo-v4+-41DAFF.svg)](https://github.com/labstack/echo)
[![Slack](https://img.shields.io/badge/go-1.11+-2A489A.svg)](https://golang.org)
[![Slack](https://img.shields.io/badge/GO111MODULE-on-brightgreen.svg)](https://golang.org)

This project exposes a cache middleware for echo with different implementations
* inmemory with [go-cache](https://github.com/patrickmn/go-cache)
* redis
* memcached


## Usage
You can see some example here https://github.com/jsdidierlaurent/echo-middleware/tree/master/cache/example

## Running tests
```bash
# Running test manually
docker-compose up -d
go test -v `go list ./... | grep -v example`

# Use makefile
make test
```
