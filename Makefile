DEFAULT: help

.PHONY: test install update help
test: ## Running tests
	@./scripts/run-tests.sh

install: ## Install deps in local cache
	@echo Installing deps in local cache ...
	@$(GOMOD) install

update: ## Update / Clean deps (and go.mod)
	@echo Updating deps and cleaning go.mod ...
	@$(GOGET)
	@$(GOMOD) tidy

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
