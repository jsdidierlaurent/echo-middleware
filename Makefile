# Go parameters
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLIST=$(GOCMD) list
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get -u

DEFAULT: help

.PHONY: all test clean install update help
all: clean test
test: ## Running tests
	@echo Starting Redis / Memcache

	@echo Runing all tests ...
	@$(GOTEST) `$(GOLIST) ./... | grep -v example`

clean: ## Clean test cache
	@echo Cleaning tests cache ...
	@$(GOCLEAN) -testcache `$(GOLIST) ./... | grep -v example`

install: ## Install deps in local cache
	@echo Installing deps in local cache ...
	@$(GOMOD) install

update: ## Update / Clean deps (and go.mod)
	@echo Updating deps and cleaning go.mod ...
	@$(GOGET)
	@$(GOMOD) tidy

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
