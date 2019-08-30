# Kubernetes Continuous Deployment Tool

[![Go Report Card](https://goreportcard.com/badge/github.com/zedge/kubecd)](https://goreportcard.com/report/github.com/zedge/kubecd)

`kubecd` is a deployment tool for Kubernetes that lets you declare in Git what should be deployed in all your
environments, manage image upgrade strategies per service, and make it so. It supports any Kubernetes installation
with some help, but has direct support for GKE or Azure, minikube and Docker, and deployment using [Helm](https://helm.sh) or plain kubectl.

Currently, `kubecd` does not implement an operator/controller, but instead integrates directly with
command-line tools. An operator is being planned, but we want to see where
[Helm 3](https://github.com/helm/community/tree/master/helm-v3/) and the
[Application CRD](https://github.com/kubernetes-sigs/application) is going first.


## Goals

 * Provide a mechanism for defining "environments", which are namespaces in clusters,
   where developers deploy their apps
 * Allow a layered config mechanism to share environment-specific values among
   Helm charts deployed in the environment (for things like Ingress domain)
 * Build as much as possible on top of existing mainstream tools (kubectl and helm)
 * Support a GitOps workflow by providing a tool with little opinion included,
   allowing you to assemble your pipeline to suit your needs.


## Configuring Environments

All the deployable environments are configured in a file typically called
`environments.yaml`. The schema for this file
[can be found here (check the `KubeCDConfig` struct)](pkg/model/model.go).

This file must contain two sections/keys, `clusters` and `environments`. Each environments maps to one
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
      - prod/releases.yaml
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
      - test/releases.yaml
    defaultValues:
      - key: "image.prefix"
        value: "gcr.io/example-com-test/"
      - key: "ingress.domain"
        value: "test.example.com"
```

Here, we have defined two environments, `test` and `prod`, each running in separate GKE clusters in
different GCP projects. We have also set some default helm chart values which will be automatically applied
to every chart deployed into those environments, so that if your chart uses the `ingress.domain` value to
construct the full Ingress host, you do not have to worry about specifying or overriding that domain part
in every single release/deployment.

## Configuring Releases

Once you have your environments defined, you need to configure what should be deployed into each of them.
This is expressed as "releases" (term borrowed from Helm).

## Installing and Running


## Contributing

### Design Document

[Can be found here.](docs/design.md)

### Running From Source

This project requires Python 3.5 and Docker.

Then you need to generated Thrift source and install a shim that runs directly from your checked out source:

    pip install -r requirements.txt
    make
    kcd --version

### Testing

To run tests, first install test dependencies:

    pip install -r requirements.txt -r requirements-dev.txt
    make test

### Making Releases

To make a new release:

 1. commit and push all changes
 2. run `bumpversion minor` (or major/patch) to update `__version__` in [`kubecd/__init__.py`](kubecd/__init__.py)
 3. run `python setup.py release` - this will make and push a Git tag, which will kick off the actual
    release process driven by Travis
