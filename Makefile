THRIFT_IMAGE=thrift:0.11

GEN_DIR=.
GEN_SUBDIR=kubecd/gen_py
KUBECD_GEN_FILE=$(GEN_SUBDIR)/ttypes.py
KUBECD_SRC_FILE=idl/github.com/zedge/kubecd/kubecd.thrift

THRIFT_GEN=docker run --rm -v $(shell pwd):$(shell pwd) -w $(shell pwd) -u $(shell id -u):$(shell id -g) $(THRIFT_IMAGE) thrift -out $(GEN_DIR) -gen py:dynamic

KCD_IMAGE=zedge/kubecd
KCD_IMAGE_TAG=latest

all: thrift-gen
.PHONY: thrift-gen test image image-push clean test release

thrift-gen: $(KUBECD_GEN_FILE)

$(KUBECD_GEN_FILE): $(KUBECD_SRC_FILE)
	$(THRIFT_GEN) $^

image:
	docker build -t $(KCD_IMAGE):$(KCD_IMAGE_TAG) .

image-push: image
	docker push $(KCD_IMAGE):$(KCD_IMAGE_TAG)

clean:
	rm -rf $(GEN_SUBDIR) $(shell find . \( -name __pycache__ -o -name \*.pyc \) -print)
	python setup.py clean

test: thrift-gen
	flake8
	pytest

release: clean test
	python setup.py release

upload:
	python setup.py clean sdist upload

build:
	@go build -o bin/kubecd ./cmd/kcd
