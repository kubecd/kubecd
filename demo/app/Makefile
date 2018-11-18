REGISTRY=eu.gcr.io/kubecd-demo
VERSION=v$(shell grep '^const Version' main.go | cut -d\" -f2)
GIT_HASH=$(shell git rev-parse --short HEAD)
APP=demo-app
IMAGE=$(REGISTRY)/$(APP)

$(APP): main.go
	go build -o $(APP) main.go

image:
	docker build -t $(APP) .

push-test: image
	docker tag $(APP) $(REGISTRY)/test-$(APP):$(GIT_HASH)
	docker push $(REGISTRY)/test-$(APP):$(GIT_HASH)

promote:
	yes | gcloud container images add-tag $(REGISTRY)/test-$(APP):$(GIT_HASH) $(REGISTRY)/prod-$(APP):$(VERSION)

.PHONY: image push-test promote
