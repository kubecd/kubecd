---
title: Kubernetes Continuous Deployment Tool
toc_title: KubeCD
---

This repository contains the `kubecd` tool for Kubernetes Continuous Deployment and cluster inventory.

## Configuring Environments

All the deployable environments are configured in a file called
`environments.yaml` by default. The schema for this file
[can be found here (check the `Environments` struct)](idl/github.com/zedge/deployments/deployments.thrift).

This file must contain two main objects, `clusters` and `environments`. Each environments maps to one
namespace in one cluster, but the environment names must be unique within this file.

Example:

```yaml
clusters:
  - name: prod-cluster
    provider:
      gke:
        project: example-com-prod
        clusterName: prod-cluster
        zone: us-central1-c
  - name: test-cluster
    provider:
      gke:
        project: example-com-test
        clusterName: test-cluster
        zone: us-central1-c

environments:
  - name: prod
    clusterName: prod-cluster
    kubeNamespace: default
    releasesFiles:
      - common/base-env.yaml
    defaultValues:
      - key: "image.prefix"
        value: "gcr.io/example-com-prod/"
      - key: "ingress.domain"
        value: "prod.example.com"
  - name: test
    clusterName: test-cluster
    kubeNamespace: default
    releasesFiles:
      - common/base-env.yaml
    defaultValues:
      - key: "image.prefix"
        value: "gcr.io/example-com-test/"
      - key: "ingress.domain"
        value: "test.example.com"
```

## Installing and Running

You can install from jfrog artifactory ([see this page](https://zedge.work/repos/zedge-dev-tools/artifactory/#python))
for how to configure Pip to fetch packages from our private repo):

    pip install kubecd
    kcd --help

## Contributing

### Design Document

[Can be found here.](docs/design.md)

### Running From Source

This project requires `pipenv`, Python 3.5 and Docker. You must install `pipenv` yourself before installing
project dependencies. See [pipenv install docs](https://docs.pipenv.org/install/) for details.

Then you need to generated Thrift source and install a shim that runs directly from your checked out source:

    pip install -e .
    make
    kcd --help

### Testing

To run tests, first install test dependencies:

    pip install -e .
    pip install -r requirements-test.txt
    make test

### Making Releases

To make a new release:

 1. commit and push all changes
 2. update the `__version__` attribute in [`kubecd/__init__.py`](kubecd/__init__.py)
 3. run `python setup.py release` - this will make a Git tag and push, which will kick off the actual
    release process driven by `cloudbuild-release.yaml`
