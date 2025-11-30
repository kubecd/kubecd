.PHONY: all build image image-multiarch image-push clean test fmt
.DEFAULT: build

IMAGE_NAME ?= kubecd/kubecd
IMAGE_TAG ?= latest

build:
	go build -ldflags "-w -s -X main.version=$$(git describe --tags)" ./cmd/kcd

image:
	docker buildx build -t $(IMAGE_NAME):$(IMAGE_TAG) --load .

image-multiarch:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME):$(IMAGE_TAG) .

image-push:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME):$(IMAGE_TAG) --push .

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
