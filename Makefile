THRIFT_IMAGE=thrift:0.11

DEPLOYMENTS_GEN_FILE=kubecd/gen_py/ttypes.py
DEPLOYMENTS_SRC_FILE=idl/github.com/zedge/kubecd/kubecd.thrift
GEN_DIR=.

THRIFT_GEN=docker run --rm -v $(shell pwd):$(shell pwd) -w $(shell pwd) -u $(shell id -u):$(shell id -g) $(THRIFT_IMAGE) thrift -out $(GEN_DIR) -gen py:dynamic

KCD_IMAGE=us.gcr.io/zedge-dev/kubecd
KCD_IMAGE_TAG=latest

all: thrift-deployments
.PHONY: thrift-deployments test image image-push

thrift-deployments: $(DEPLOYMENTS_GEN_FILE)

$(DEPLOYMENTS_GEN_FILE): $(DEPLOYMENTS_SRC_FILE)
	@mkdir -p $(GEN_DIR)
	$(THRIFT_GEN) $^

test:
	PYTHONPATH=$$(pwd) pipenv run pytest

image:
	docker build -t $(KCD_IMAGE):$(KCD_IMAGE_TAG) .

image-push: image
	docker push $(KCD_IMAGE):$(KCD_IMAGE_TAG)
