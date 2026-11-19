{{- define "name" -}}
gardener-extension-diki
{{- end -}}

{{- define "leaderelection.id" -}}
extension-diki-leader-election
{{- end -}}

{{-  define "image" -}}
  {{- if .Values.image.ref -}}
  {{ .Values.image.ref }}
  {{- else -}}
  {{- if hasPrefix "sha256:" .Values.image.tag }}
  {{- printf "%s@%s" .Values.image.repository .Values.image.tag }}
  {{- else }}
  {{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
  {{- end }}
  {{- end -}}
{{- end }}

{{- define "config" -}}
apiVersion: config.diki.extensions.gardener.cloud/v1alpha1
kind: Configuration
{{- if and .Values.dikiServiceConfig .Values.dikiServiceConfig.baseDikiConfig }}
baseDikiConfig: |
{{ .Values.dikiServiceConfig.baseDikiConfig | indent 2 }}
{{- end }}
{{- if and .Values.dikiServiceConfig .Values.dikiServiceConfig.defaultScheduledScan }}
defaultScheduledScan:
  {{- with .Values.dikiServiceConfig.defaultScheduledScan }}
  schedule: {{ .schedule | quote }}
  {{- if .successfulScansHistoryLimit }}
  successfulScansHistoryLimit: {{ .successfulScansHistoryLimit }}
  {{- end }}
  {{- if .failedScansHistoryLimit }}
  failedScansHistoryLimit: {{ .failedScansHistoryLimit }}
  {{- end }}
  {{- if .rulesets }}
  rulesets:
  {{- range .rulesets }}
  - id: {{ .id | quote }}
    version: {{ .version | quote }}
  {{- end }}
  {{- end }}
  {{- end }}
{{- end }}
{{- end -}}
