---
title: Kubernetes Continuous Deployment - DesignDoc
toc_title: KubeCD
---

## Goals

We aim to provide the following:

1. A way to declare what should be running in a Kubernetes cluster using [GitOps](#gitops) principles, and
   tools to make it so.
1. A low-boilerplate continuous delivery mechanism for IT and developers, that:
   * ...is easy to getting started with, and is maintainable over time.
   * ...is agnostic of CI/build servers
   * ...allows for sharing properties among services in the same cluster/environment (such as external domain).
   * ...can be used for in-house as well as third-party services
1. Separation of (at least) test and production environments, with the option of separate access control
   per environment.
1. Multi-vendor support by design
   * we will have to adapt to certain aspects of each cloud vendor, but avoid locking ourselves
     to vendor-specific solutions without good reason

### Non-Goals and Limitations

* Provisioning clusters is not (currently) in scope. Input wanted.
* Kubernetes only
* Not aiming to support older Kubernetes versions (requiring 1.8 is reasonable)

## Overall Vision

A rough sketch of the desired development and deployment workflow:

![Architecture](./architecture.svg)
*[Drawing source link](https://docs.google.com/drawings/d/1QCnoAhU6X6Y4po_PCCzBFp20VNoAHHYh8gJvUh89egk)*

➊ Developers work on individual repos (for example `backend-api-browse.git`), using pull requests, tags and so on.

❷ A build server assembles Docker images from these, and attaches the relevant Git metadata such as repository,
  branch, commit/tag and so on. The image needs to pass its tests before being pushed. Any build server could
  be used here in theory, allowing teams to choose what works best for them.

➌ The Docker image is pushed to a registry (`us.gcr.io/zedge-dev/backend-api-browse` here).

➍ A build server is notified of the newly pushed image (webhook or pubsub)
 
➎ A build for `deployments.git` is started, scans upgrade triggers for every release, bumps the image
  tag for deployments that want this image, and creates a PR for this change.

➏ PR is merged (either after being reviewed by a human, or automatically), and `deployments.git` master branch is
  updated with the new image tag.

➐ Within each cluster, a deployment agent regularly pulls `deployments.git`

➑ If changes are detected, deployments are re-applied

➒ Slack notification (we could have more of these underway)

## Environments

An "environment" in this context is a namespace within a Kubernetes cluster.

A config file describes all of the reachable environments. For example:

```yaml
environments:
  - name: analytics
    provider:
      gke:
        project: zedge-prod
        clusterName: analytics
        zone: us-central1-b
    kubeNamespace: default
    releases:
      - ./base/releases.yaml
      - ./analytics/releases.yaml
    defaultValues:
      - key: "ingress.domain"
        value: "zedge.io"
  - name: airflow
    provider:
      gke:
        project: zedge-prod
        clusterName: analytics
        zone: us-central1-b
    kubeNamespace: airflow
    releases:
      - ./analytics/airflow-releases.yaml
    defaultValues:
      - key: "ingress.domain"
        value: "zedge.io"
```

Here, there are two environments, both of them in the zedge-prod analytics cluster, but in separate namespaces.
The `analytics` environment includes two release files, while the `airflow` environment only refers to one.

The `provider` and `kubeNamespace` fields should be self-explanatory.

The `releases` field specifies a list of "release files", which are parameterized Helm charts, and will be
explained below.

The `defaultValues` field specifies a list of Helm values that will be used for every chart/release
deployed into the environment.

The full schema for this file can be found in the `Environments` struct in
[`kubecd.thrift`](idl/github.com/zedge/kubecd/kubecd.thrift).

## Releases

The term "release" here is borrowed from Helm. Helm, somewhat confusingly, uses this term
to describe an [installation of a chart](https://github.com/kubernetes/helm/blob/master/docs/glossary.md#release).

In KubeCD, a "release" is also an instance of a Helm chart, but it comes with some extra attributes:

 * value overrides for the chart (either as a yaml file, a list of key/value pairs, or both)
 * an upgrade trigger

### Standardized Values

Helm's values files are not typed, so you can have any structure you want in there. However, if we
standardize a few values, a lot of opportunities become available.

KubeCD itself needs to know about two or three values to check for image

| Value name | Usage |
| --- | --- |
| `image.repository` | The image repository, which is Docker's term for the registry host plus "path" of the image, for example `gcr.io/project/image` |
| `image.tag` | The "tag" portion of the chart's main image, for example `2.0` |
| `image.prefix` | (Optional) if present, will be prepended to `image.repository`. This value is useful for cases where images in different environments reside in different Docker registries (such as with GCR and environments in different GCP projects) |

For sub-charts, the same pattern applies with the sub-chart name prefixed, for example
`zookeeper.image.repository`.

#### Semantic Versioning

We wish to support semantic versioning for images that follow it. In this case we will recognize tags using these
formats: `X`, `X.Y`, `X.Y.Z`, `vX`, `vX.Y`, `vX.Y.Z`. The `v` prefix is common enough with open source software to
justify making an exception for it.

### Examples

```yaml
resourceFiles:
  - storage-fast.yaml

releases:

  - name: clickhouse
    chart:
      dir: ./charts/clickhouse
    values:
      - key: image.tag
        value: "1.1"

  - name: kafkaesque
    chart:
      dir: ./charts/kafkaesque
    triggers:
      - image:
          track: Newest
    
  - name: kafka
    chart:
      dir: ../charts/kafka
    valuesFile: ./kafka/values-prod.yaml
    values:
      - key: image.tag
        value: "4.0"
    triggers:
      - image:
          tagValue: "image.tag"
          repoValue: "image.repository"
          track: PatchLevel
      - image:
          tagValue: "zookeeper.image.tag"
          repoValue: "zookeeper.image.repository"
          track: PatchLevel
```

This file serves two roles: it has values to describe exactly which image should be
used for each release, and optionally a trigger for detecting new versions of the image.

It also has a list of `resourceFiles` which are straight Kubernetes resource files, with no fancy
substitutions or upgrade triggers happening. These are meant for simple stuff such as storage classes,
cluster roles and role bindings, shared config maps that are generated by other systems, and so on. A
major caveat with resources created this way is that they won't be automatically removed if you remove
them from the release file. If this is a problem convert it to a Helm chart.

The three releases included in the example show three different types of deployments we expect
to be common: a chart locked to a specific third-party image version (clickhouse), a chart tracking
patch level updates to third-party images (kafka) and finally a chart simply using the newest image
available (kafkaesque). We'll go through the examples below.

Let's go through each of the examples:

#### `clickhouse` release:

```yaml
  - name: clickhouse
    chart:
      dir: ./charts/clickhouse
    values:
      - key: image.tag
        value: "1.1"
```

This is a very simple release that sets the `image.tag` value but does not attempt to track image
updates.

#### `kafkaesque` release:

```yaml
  - name: kafkaesque
    chart:
      dir: ./charts/backend-api-browse
    values:
      - key: image.repository
        value: "gcr.io/zedge-prod/backend-api-browse"
      - key: image.tag
        value: "v20170911-0926"
    triggers:
      - image:
          track: Newest
```

Again, we apply a chart from a local directory, but now there'a also an upgrade trigger. More on upgrade triggers
[below](#upgrade-triggers).

#### `nginx-ingress` release:

```yaml
  - name: ingress
    chart:
      reference: stable/nginx-ingress
      version: 0.9.5
    values:
      - key: controller.service.loadBalancerIP
        valueFrom:
          gceResource:
            address:
              name: analytics-ingress
              isGlobal: false
```

This example shows how to use public charts from official repos. We have chosen to call these "chart references"
because some Helm documentation used that term. For reproducibility, it's important to pin the chart version,
so the `version` field needs to be required for chart references.

Another feature shown off in this example is how to fill in data from the underlying infrastructure. Here, we
set the `controller.service.loadBalancerIP` value to the IP address of the assumed-to-be-preallocated Google Cloud
Platform IP address named `analytics-ingress` in the same project/region as the environment/cluster.
(If `isGlobal` had been `true`, it would be using a global (anycast) IP, not a regional one.)

### Upgrade Triggers

An "upgrade trigger" describes a means to react to events like a new image tag. Each release may specify
which repositories to watch for tags, and which tags should be considered.

Triggers are not used while deploying a release, instead they declare how the system should look for
new versions during an upgrade check, and which versions this release wants to subscribe to.

This can be just informative for engineers, but the idea is that an automated update of the tag value happens,
as shown in ➎ in [the architecture diagram](#overall-vision) section.

For example:

```yaml
    values:
      - key: "image"
        value: "mysql"
      - key: "imageTag"
        value: "5.7.14"
    triggers:
      - image:
          repoValue: "image"
          tagValue: "imageTag"
          track: PatchLevel
```

Here, the upgrade trigger will kick in for patch level updates within the same major.minor version.
In other words, if a `5.7.15` tag is found for the image, this is considered an upgrade, but a `5.8.0` tag
is not.

The supported values for `track:` are:

 * `PatchLevel` - upgrade on patch level updates only
 * `MinorVersion` - upgrade on minor version or patch level updates
 * `MajorVersion` - upgrade on minor, major or patch level version updates
 * `Newest` - upgrade to whatever is the newest (as in most recently created) tag

The `repoValue` and `tagValue` fields refer to the Helm values containing the image repository and tag
respectively. If omitted, these default to `"image.repository"` and `"image.tag"`, which are the standard
KubeCD values.

It is possible to watch for updates in more than one image. This can be useful if a chart consists of multiple
workloads, as seen in the `kafka` example above.

#### Auto Upgrades

Longer-term, the upgrade checks should patch the same file it got its triggers from with new tag values,
and either create a PR or just commit to master. We probably want to get some hands-on experience with
upgrade checks before taking this step.

Auto-patching needs to be done in a way that preserves comments and field order, for example using
[`ruamel.yaml`](http://yaml.readthedocs.io/en/latest/overview.html) in Python.

An alternative to directly patching the `releases.yaml` file could be to have an auto-generated
lock file with tag values on the side. This would remove the comment preservation / field order requirement.

## Apply Agent

The "apply agent" is a controller loop running in each cluster/environment.
It will do an "apply" of the entire environment either at a fixed schedule (for example every 5 minutes)
or triggered by webhooks.

## Roadmap

The system should be implemented in stages. Here's a proposal, focusing on getting something up and running
quickly and then automating, optimizing and securing the system over time:

1. "Apply" stage: developers run an "apply" command manually, using their own Kubernetes credentials. This is
   still a fairly manual process, but still a big improvement over what we have today, especially for IT.
2. "Apply from cron" stage: using a service account, automatically run (1) every 1-5 minutes. All changes
   are still made manually.
3. "Upgrade check" stage: the upgrade checks are implemented, but auto-patching is not implemented. 
   Developers would manually patch image tag values based on the upgrade check output.
4. "Auto upgrade" stage: implement the CI job creating PRs based on upgrade check changes. Will need
   a GitHub bot to implement per-environment access control if we go for one big deployments repo.
5. "Cluster agent" stage: implement the agent/operator running within each cluster or environment,
   pulling the deployments repo and applying changes. At this point we will no longer need a super
   cron job with credentials for every cluster, which is better for both maintenance and security.
6. "Webhook all the things" stage: convert image tag polling to using webhooks or pubsub, to
   speed up deployments and reduce overhead. Also convert cluster agents to use webhooks instead
   of polling.


## TODO

Things we have not covered yet, but probably should...

### TODO: Secrets

We still don't have a solution for automatically making secrets from Vault available to Kubernetes apps.

One approach we could explore is to declare a way of populating Kubernetes secrets from Vault in this system,
keeping the secrets themselves out of the `deployments.git` repo, using Vault path references instead.

Please use issue #3 is for discussing secrets.

### TODO: Config

Config that is maintained separately from the applications needs to be projected into the deployments repo
somehow.

### TODO: Provisioning integration

The `environments.yaml` file refers to cloud resources, but we have not defined how these are created.
This is consciously omitted right now to keep the scope down, but will have to be addressed.

Please use issue #4 for discussing provisioning.

### TODO: Blue/Green / Staged / Canary Rollouts

The proposed solution has no built-in mechanisms for blue/green or canary rollouts.
This should probably be considered separately, after we have the basic CD mechanisms in place.

## References

### GitOps

"GitOps" is a newer coinage of the concept "infrastructure as code"
that has been popularized lately especially by WeaveWorks. It
describes operation of server-side infrastructure using Git as the
source of truth. This way one can re-use existing tools for version
control, history, reviews, rollbacks and so on.

Some references:

 * [WeaveWorks Blog](https://www.weave.works/blog/gitops-operations-by-pull-request)
 * [KubeCon 2017 keynote by Kelsey Hightower](https://www.youtube.com/watch?v=07jq-5VbBVQ)

### Helm

[Helm](https://helm.sh) is the standard Kubernetes "package
manager". It can be used to package multiple related resources in a
single "chart", with a parameterization layer and support for running
multiple instances of the same package in a cluster.

Helm uses
[go templates](https://github.com/kubernetes/helm/blob/master/docs/charts.md#templates-and-values)
on top of Kubernetes YAML resource files. Using templates with a
whitespace-sensitive format like YAML can be tricky, and something
we should not burden every developer to have to deal with. (A notable tool to keep an eye on
in this space is [ksonnet](https://github.com/ksonnet/ksonnet).)

Helm's general solution to this is to offer "values" that the user can
customize, through an arbitrary YAML schema defined in each chart's
"values file".

We can simplify the developer experience by standardizing the
structure of the values file for in-house services, so that when
developers learn how to configure resource requirements, external host
names and path and son on, this learning can be applied to any
in-house Zedge service.

### Other Tools / References

 * https://github.com/weaveworks/flux
 * https://coreos.com/blog/introducing-operators.html
 * https://github.com/GoogleCloudPlatform/skaffold
 * https://github.com/keel-hq/keel - implements continuous deployments, but lacks a mechanism
   for re-creating a cluster
 * https://github.com/GoogleCloudPlatform/continuous-deployment-on-kubernetes
 * https://github.com/kelseyhightower/pipeline
 * https://github.com/ksonnet/ksonnet
 