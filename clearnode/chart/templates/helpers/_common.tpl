{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the component.
*/}}
{{- define "clearnode.common.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully common name.
If release name contains chart name it will be used as a full name.
*/}}
{{- define "clearnode.common.fullname" -}}
{{- if .Values.prefixOverride }}
{{- printf "%s-%s" .Values.prefixOverride .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- if contains .Chart.Name .Release.Name }}
{{- print .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Common Selector labels
*/}}
{{- define "clearnode.common.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "clearnode.common.labels" -}}
helm.sh/chart: {{ include "clearnode.common.chart" . }}
{{ include "clearnode.common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.extraLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "clearnode.common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Returns common image pull secrets
*/}}
{{- define "clearnode.common.imagePullSecrets" -}}
{{- with .Values.imagePullSecret }}
imagePullSecrets:   
- name: {{ . }}
{{- end }}
{{- end }}

{{/*
Returns common environment variables
*/}}
{{- define "clearnode.common.env" -}}
- name: CLEARNODE_LOG_LEVEL
  value: {{ .Values.config.logLevel }}
- name: CLEARNODE_CONFIG_DIR_PATH
  value: /app/config
- name: CLEARNODE_MODE
  value: production
- name: DATABASE_DRIVER
  value: {{ .Values.config.database.driver }}
{{- range $key, $value := .Values.config.extraEnvs }}
- name: {{ $key | upper }}
  value: {{ $value | print | quote }}
{{- end }}
{{- end }}

{{/*
Returns common node selector labels
*/}}
{{- define "clearnode.common.nodeSelectorLabels" -}}
{{- with .Values.nodeSelector }}
nodeSelector:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Returns common tolerations
*/}}
{{- define "clearnode.common.tolerations" -}}
{{- with .Values.tolerations }}
tolerations:
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Returns common pod's affinity
*/}}
{{- define "clearnode.common.affinity" -}}
{{- with .Values.affinity }}
affinity:
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end }}
