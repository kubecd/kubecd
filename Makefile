.PHONY: all build build-all image image-push clean test fmt
.DEFAULT: build

GOARCH ?= $(shell go env GOARCH)

build:
	GOARCH=$(GOARCH) go build -ldflags "-w -s -X main.version=$$(git describe --tags)" ./cmd/kcd

build-all:
	GOARCH=amd64 go build -ldflags "-w -s -X main.version=$$(git describe --tags)" -o kcd-linux-amd64 ./cmd/kcd
	GOARCH=arm64 go build -ldflags "-w -s -X main.version=$$(git describe --tags)" -o kcd-linux-arm64 ./cmd/kcd

clean:
	go clean ./...

fmt:
	@if ! test -z `gofmt -s -l cmd pkg`; then \
	  echo "gofmt failed, please fix by running:"; \
	  echo "gofmt -w "`gofmt -s -l cmd pkg`; \
	  exit 1; \
	fi >&2

test: fmt
	go vet ./...
	go test ./...

clean-workspace:
	@(test "$(git status --short)" = '' && git diff --quiet) || { \
	  echo "Workspace not clean!"; \
	  exit 1; \
	}

release: clean-workspace
	set -e; \
	tag=$$(go run ./cmd/bumpversion/ $(BUMP)); \
	echo "Next tag: $$tag"; \
	git tag $$tag
	git push --tags
	git push
