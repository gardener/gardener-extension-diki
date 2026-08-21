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
{{- if and .Values.dikiServiceConfig .Values.dikiServiceConfig.baseDikiOptions }}
baseDikiOptions:
  data: |
{{ .Values.dikiServiceConfig.baseDikiOptions.data | indent 4 }}
{{- end }}
{{- end -}}
