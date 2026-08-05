# gardener-extension-diki

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/gardener-extension-diki)](https://api.reuse.software/info/github.com/gardener/gardener-extension-diki)

## Overview

`gardener-extension-diki` is a [Gardener](https://gardener.cloud/) extension that deploys the
[diki-operator](https://github.com/gardener/diki-operator) into shoot cluster control planes,
enabling declarative compliance scanning through native Kubernetes resources.

The extension follows the standard Gardener extension contract as described in
[GEP-0063](https://github.com/gardener/enhancements/pull/64).

## Features

- Deploys `diki-operator` into shoot control planes on seeds
- Allows users to run on-demand compliance scans via `ComplianceScan` custom resources
- Supports recurring scans via `ScheduledComplianceScan` custom resources
- Provides configurable report outputs via `ReportOutput` custom resources
- Full lifecycle management (reconcile, delete, migrate, restore, force-delete)

## Documentation

Please find the documentation in the [`/docs`](./docs) directory.

## Feedback and Support

Feedback and contributions are always welcome. Please report bugs or suggestions as
[GitHub issues](https://github.com/gardener/gardener-extension-diki/issues) or join
us on [Slack](https://gardener-cloud.slack.com/) (join the workspace [here](https://gardener.cloud/community/)).
