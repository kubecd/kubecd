KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

.PHONY: all build image image-push clean test fmt
.DEFAULT: build

build:
	go build ./cmd/kcd

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

#release: clean test
