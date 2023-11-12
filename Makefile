SHELL=/bin/bash

.PHONY: setup
setup:
	pre-commit install

.PHONY: lint
lint:
	pre-commit run --all-files

.PHONY: build
build:
	go build -o pkgforbid cmd/main.go
