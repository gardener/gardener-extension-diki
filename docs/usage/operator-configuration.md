# Operator Configuration

The extension controller is configured via a `Configuration` object (passed via `--config` flag, usually mounted from a Helm chart value):

```yaml
apiVersion: diki.extension/v1alpha1
kind: Configuration
# Optional: base diki options applied to all compliance scans (providers, rulesets, rule exceptions).
baseDikiConfig: |
  providers:
  - id: managedk8s
    name: "Managed Kubernetes"
    rulesets:
    - id: disa-kubernetes-stig
      version: v2r6
      ruleOptions:
      - ruleID: "242417"
        args:
          acceptedPods:
          - labelSelector:
              matchLabels:
                resources.gardener.cloud/managed-by: gardener
            justification: "Pods managed by Gardener are not considered as user pods."
# Optional: inject a default ScheduledComplianceScan into every shoot that has the extension enabled.
# The hour and minute fields of the schedule are replaced with values derived from the shoot's
# maintenance time window begin time; the day-of-month, month, and day-of-week fields are
# preserved as configured. If no maintenance window is configured, the schedule is used as-is.
defaultScheduledScan:
  # schedule is a standard 5-field cron expression. Examples:
  #   daily:        "0 0 * * *"
  #   weekly:       "0 0 * * 0"
  #   twice/month:  "0 0 1,15 * *"
  schedule: "0 0 * * *"
  successfulScansHistoryLimit: 3
  failedScansHistoryLimit: 1
  rulesets:
  - id: disa-kubernetes-stig
    version: v2r6
```

> [!IMPORTANT]
> When `defaultScheduledScan` is configured, the extension's `ControllerRegistration` (or `Extension` resource for Gardener operator deployments) **must** set `lifecycle.reconcile: AfterWorker`. This ensures the extension is only reconciled after the shoot's worker nodes are ready, which is required for compliance scans to run successfully.

When `defaultScheduledScan` is set, a `ScheduledComplianceScan` named `default` is injected into every shoot cluster that has the extension enabled. If the shoot has a maintenance time window configured, the hour and minute fields of `schedule` are replaced with values derived from `spec.maintenance.timeWindow.begin` (e.g. `220000+0000` → hour=22, minute=0), while the day-of-month, month, and day-of-week fields remain as configured by the operator. If the shoot has no maintenance window, `schedule` is used as-is.
