THRIFT_IMAGE=thrift:0.11

GEN_DIR=.
GEN_SUBDIR=kubecd/gen_py
KUBECD_GEN_FILE=$(GEN_SUBDIR)/ttypes.py
KUBECD_SRC_FILE=idl/github.com/zedge/kubecd/kubecd.thrift

KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

all: build

build:
	GOOS=darwin GOARCH=amd64 go build -o kcd-darwin-amd64 ./cmd/kcd
	GOOS=linux GOARCH=amd64 go build -o kcd-linux-amd64 ./cmd/kcd

.PHONY: thrift-gen test image image-push clean test release build

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

