# K8s Network Prober

<a href="https://gitpod.io/#https://github.com/bigmikes/k8s-network-prober">
  <img
    src="https://img.shields.io/badge/Contribute%20with-Gitpod-908a85?logo=gitpod"
    alt="Contribute with Gitpod"
  />
</a>

## Overview

K8s Network Prober allows you to monitor the latency between every Pod of your Kubernetes cluster. The Network Prober container can be deployed with its [Operator](https://github.com/bigmikes/k8s-network-prober-operator). Every instance of Network Prober export Prometheus metrics in form of Histogram Vector.

## Kubernetes Operator

Follow the Operator's [README](https://github.com/bigmikes/k8s-network-prober-operator) to deploy the application on your K8s cluster. 

## Docker Build and Publish

Use the provided Makefile to build the container image and publish it to your repository.

```bash
make docker-build IMG=bigmikes/net-prober:test-version-v9
make docker-publish IMG=bigmikes/net-prober:test-version-v9
```