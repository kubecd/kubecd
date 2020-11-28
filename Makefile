.PHONY: all build image image-push clean test fmt
.DEFAULT: build

build:
	go build -ldflags "-w -s -X main.version=$$(git describe --tags)" ./cmd/kcd

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
