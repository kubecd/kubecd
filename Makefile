KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

.PHONY: all build image image-push clean test fmt
.DEFAULT: build

build:
	go build -ldflags "-w -s -X main.version=$$(git describe --tags)" ./cmd/kcd

image:
	docker build -t $(KCD_IMAGE):$(KCD_IMAGE_TAG) .

image-push: image
	docker push $(KCD_IMAGE):$(KCD_IMAGE_TAG)

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
