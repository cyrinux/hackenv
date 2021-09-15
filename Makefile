SCRIPTS := $(wildcard ./bin/*)

BASEDIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: build
build:
	mkdir -p ./out
	go build -ldflags="-w -s" -o ./out/hackenv ./cmd/hackenv

.PHONY: test
test: lint_scripts
	stdout=$$(gofmt -l . 2>&1); \
	if [ "$$stdout" ]; then \
		exit 1; \
	fi
	go vet ./...
	gocyclo -over 10 $(shell find . -iname '*.go' -type f)
	staticcheck ./...
	go test -v -cover ./...

.PHONY: setup
setup:
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: lint_scripts
lint_scripts:
	shellcheck ${SCRIPTS}

.PHONY: install_scripts
install_scripts:
	ln -i -s -r ${BASEDIR}/${SCRIPTS} ${HOME}/bin/
