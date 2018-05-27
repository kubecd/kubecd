---
title: KubeCD Modules and Components
---

## Background

Currently KubeCD uses resource files and (Helm) releases as its main deployment units.
This has the advantage of being simple to map to `kubectl` and `helm` commands, but also
some disadvantages:

 * Deployment order is maintained within resource files and Helm releases separately,
   with Helm after resources. This means that if a resource file depends on a CRD
   that is created through Helm, there is no robust way of automating.
 * It's hard to supporting more deployment mechanisms / manifest formats (such as
   [ksonnet](https://github.com/ksonnet/ksonnet/))
 * You cannot logically group multiple resource files (you are forced to stuff everything
   into one file)

## Module

This is what was previously known as `Releases` - a collection of API objects and Helm
releases defined in a single YAML file.

## Component

This is either a collection of API objects (kubectl files) or a Helm release.
