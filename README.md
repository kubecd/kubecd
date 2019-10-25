# Kubernetes Continuous Deployment Tool

![](https://github.com/zedge/kubecd/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zedge/kubecd)](https://goreportcard.com/report/github.com/zedge/kubecd)
![](https://img.shields.io/github/v/release/zedge/kubecd.svg)

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
helmRepos:
  - name: stable
    url: https://kubernetes-charts.storage.googleapis.com/

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

A "releases" file contains a list of releases, for example:

```yaml
releases:
  - name: ingress
    chart:
      reference: stable/nginx-ingress
      version: 1.15.0
    valuesFile: values-ingress.yaml
```

See more examples here: [releases-common.yaml](demo/releases-common.yaml),
[releases-prod.yaml](demo/releases-prod.yaml), [releases-test.yaml](demo/releases-test.yaml).

## Installing and Running

To produce a `kcd` binary:

```
make build
```

## Contributing

When submitting PRs, please ensure your code passes `gofmt`, `go vet` and `go test`.

For bigger changes, please [create an issue](https://github.com/zedge/kubecd/issues/new) proposing the
change in advance.


### Design Document

[Can be found here.](docs/design.md)
