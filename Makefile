KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

.PHONY: all build image image-push clean test release upload

all: build

build:
	GOOS=darwin GOARCH=amd64 go build -o kcd-darwin-amd64 ./cmd/kcd
	GOOS=linux GOARCH=amd64 go build -o kcd-linux-amd64 ./cmd/kcd


image:
	docker build -t $(KCD_IMAGE):$(KCD_IMAGE_TAG) .

image-push: image
	docker push $(KCD_IMAGE):$(KCD_IMAGE_TAG)

clean:
	go clean

test:
	go test

#release: clean test

#upload:
