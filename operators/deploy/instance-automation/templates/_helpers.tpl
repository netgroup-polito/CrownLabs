{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "instance-automation.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "instance-automation.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "instance-automation.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
The version of the application to be deployed
*/}}
{{- define "instance-automation.version" -}}
{{- if .Values.global }}
{{- .Values.image.tag | default .Values.global.version | default .Chart.AppVersion }}
{{- else }}
{{- .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "instance-automation.labels" -}}
helm.sh/chart: {{ include "instance-automation.chart" . }}
{{ include "instance-automation.selectorLabels" . }}
app.kubernetes.io/version: {{ include "instance-automation.version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "instance-automation.selectorLabels" -}}
app.kubernetes.io/name: {{ include "instance-automation.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Metrics selector additional labels
*/}}
{{- define "instance-automation.metricsAdditionalLabels" -}}
app.kubernetes.io/component: metrics
{{- end }}

{{/*
The tag to be used for sidecar containers images
*/}}
{{- define "instance-automation.containerEnvironmentSidecarsTag" -}}
{{- .Values.configurations.containerEnvironmentOptions.tag | default ( include "instance-automation.version" . ) }}
{{- end }}

{{/*
The tag to be used for image exporter container for VM snapshots
*/}}
{{- define "instance-automation.containerExportImageTag" -}}
{{- .Values.configurations.containerVmSnapshots.exportImageTag | default ( include "instance-automation.version" . ) }}
{{- end }}