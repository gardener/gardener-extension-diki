---
title: Deploying Diki Extension Locally
description: Learn how to set up a local development environment
---

# Deploying Diki Extension Locally

## Prerequisites

- Make sure that you have a running local Gardener setup. The steps to complete this can be found in the [Deploying Gardener Locally guide](https://github.com/gardener/gardener/blob/master/docs/deployment/getting_started_locally.md).

> [!TIP]
> Ensure that the locally used Gardener version matches the version specified by the `github.com/gardener/gardener` dependency.
> The extension's local setup must run successfully against a local Gardener setup at the version referenced by this dependency.

> [!NOTE]
> The location of the Gardener project is expected to be under the same root (e.g. `~/go/src/github.com/gardener/`). If this is not the case, the location of Gardener project should be specified in `GARDENER_REPO_ROOT` environment variable:
> ```bash
> export GARDENER_REPO_ROOT="<path_to_gardener_project>"
> ```

## Setting up the Diki Extension

```bash
make extension-up
```

The corresponding make target will build the diki extension container image and the extension chart OCI artifact. Then, the container image and the OCI artifact are pushed into the default skaffold registry (i.e. `registry.local.gardener.cloud:5001`). Next, the diki `Extension.operator.gardener.cloud` resource is deployed into the KinD cluster. Based on this resource the gardener-operator will deploy the diki ControllerDeployment and ControllerRegistration resources.

### Development Mode

For iterative development with automatic rebuilds on file changes, use:

```bash
make extension-dev
```

This starts skaffold in dev mode with manual trigger and no cleanup, allowing you to rebuild and redeploy by pressing any key in the terminal.

## Creating a Shoot Cluster

> [!NOTE]
> Make sure that your `KUBECONFIG` environment variable is targeting the virtual Garden cluster (i.e. `<path_to_gardener_project>/dev-setup/kubeconfigs/virtual-garden/kubeconfig`).

Once the above step is completed, you can create a Shoot cluster with the `diki` extension enabled in its specification.

## Tearing Down the Development Environment

To tear down the development environment, delete the Shoot cluster or disable the `diki` extension in the Shoot's specification. When the extension is not used by the Shoot anymore, you can run:

```bash
make extension-down
```

The corresponding make target will delete the `Extension.operator.gardener.cloud` resource. Consequently, the gardener-operator will delete the diki ControllerDeployment and ControllerRegistration resources.
