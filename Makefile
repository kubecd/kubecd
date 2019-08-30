KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

.PHONY: all build image image-push clean test release upload

all: build

build:
	go build ./cmd/kcd

image:
	docker build -t $(KCD_IMAGE):$(KCD_IMAGE_TAG) .

image-push: image
	docker push $(KCD_IMAGE):$(KCD_IMAGE_TAG)

clean:
	go clean ./...

test:
	test -z `gofmt -s -l cmd pkg`
	go vet ./...
	go test ./...

#release: clean test

#upload:
