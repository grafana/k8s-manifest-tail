{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-manifest-tail.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "k8s-manifest-tail.fullname" -}}
{{- default .Chart.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-manifest-tail.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{ include "k8s-manifest-tail.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-manifest-tail.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-manifest-tail.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name
*/}}
{{- define "k8s-manifest-tail.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "k8s-manifest-tail.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image reference
*/}}
{{- define "k8s-manifest-tail.image" -}}
{{- $registry := .Values.global.image.registry | default .Values.image.registry }}
{{- if .Values.image.digest }}
{{- printf "%s/%s@%s" $registry .Values.image.repository .Values.image.digest }}
{{- else }}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry .Values.image.repository $tag }}
{{- end }}
{{- end }}

{{/*
Image pull secrets - merges global and chart-level pull secrets
*/}}
{{- define "k8s-manifest-tail.imagePullSecrets" -}}
{{- $secrets := concat (.Values.global.image.pullSecrets | default (list)) (.Values.image.pullSecrets | default (list)) | uniq }}
{{- if $secrets }}
imagePullSecrets:
  {{- range $secrets }}
  - name: {{ . }}
  {{- end }}
{{- end }}
{{- end }}
