.DEFAULT_GOAL := help

ROOT := $(patsubst %/,%,$(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
BIN := $(ROOT)/bin
DOTFILES := $(ROOT)/.dotfiles

.PHONY: tools
tools: ## install tools
	go install github.com/goreleaser/goreleaser@latest



.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
