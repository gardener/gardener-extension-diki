# Configuring the Diki Extension for Shoot Clusters

## Overview

The `diki` extension deploys the [diki-operator](https://github.com/gardener/diki-operator/tree/v0.1.0) into the Shoot cluster's control plane on the Seed. The diki-operator orchestrates compliance scanning for Kubernetes clusters using the [Diki](https://github.com/gardener/diki) compliance checker.

Once enabled, Shoot cluster users can create `ComplianceScan`, `ScheduledComplianceScan`, and `ReportOutput` resources in the Shoot cluster to run compliance checks against supported rulesets (e.g. DISA Kubernetes STIG).

## Shoot Configuration

### Enabling the Extension

The extension is not active by default and must be explicitly enabled per Shoot. Add a `diki` entry to `spec.extensions` in your Shoot resource:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: Shoot
metadata:
  name: my-shoot
  namespace: garden-my-project
spec:
  extensions:
  - type: diki
  # ...
```

The extension does not require any `providerConfig`. Once the Shoot is reconciled, the diki-operator is deployed and the necessary CRDs (`ComplianceScan`, `ScheduledComplianceScan`, `ReportOutput`) are available in the Shoot cluster.

### Disabling the Extension

To disable the extension, remove the `diki` entry from `spec.extensions`. The extension cleans up all deployed resources. Note that deleting the CRDs cascades to all custom resources (compliance scans, scheduled scans, report outputs) in the Shoot cluster.

## Running Compliance Scans

After enabling the extension, you can create compliance scan resources in the Shoot cluster. For full details on the custom resources and their fields, refer to the [diki-operator documentation](https://github.com/gardener/diki-operator/tree/v0.1.0).

### ReportOutput

A `ReportOutput` defines where compliance scan reports are stored. Create it before referencing it in a `ComplianceScan`:

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: compliance-scan-report
spec:
  output:
    configMap:
      namespace: kube-system
      namePrefix: compliance-scan-report-
```

### ComplianceScan

A `ComplianceScan` represents a single compliance scan run. Its spec is immutable after creation:

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ComplianceScan
metadata:
  name: example-compliancescan
spec:
  rulesets:
  - id: disa-kubernetes-stig
    version: v2r6
    options:
      ruleset:
        configMapRef:
          name: diki-options
          namespace: kube-system
  outputs:
  - name: compliance-scan-report
```

### ScheduledComplianceScan

A `ScheduledComplianceScan` defines a cron schedule for recurring compliance scans:

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ScheduledComplianceScan
metadata:
  name: weekly-scan
spec:
  schedule: "0 0 * * 0"
  successfulScansHistoryLimit: 3
  failedScansHistoryLimit: 1
  scanTemplate:
    spec:
      rulesets:
      - id: disa-kubernetes-stig
        version: v2r6
      outputs:
      - name: compliance-scan-report
```
