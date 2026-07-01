{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bastion.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bastion.fullname" -}}
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
The version of the application to be deployed
*/}}
{{- define "bastion.version" -}}
{{- if .Values.global }}
{{- .Values.image.tag | default .Values.global.version | default .Chart.AppVersion }}
{{- else }}
{{- .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bastion.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bastion.labels" -}}
helm.sh/chart: {{ include "bastion.chart" . }}
{{ include "bastion.selectorLabels" . }}
app.kubernetes.io/version: {{ include "bastion.version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "webssh.labels" -}}
helm.sh/chart: {{ include "bastion.chart" . }}
{{ include "webssh.selectorLabels" . }}
app.kubernetes.io/version: {{ include "bastion.version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: webssh
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bastion.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bastion.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "webssh.selectorLabels" -}}
app.kubernetes.io/name: webssh
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Metrics selector additional labels
*/}}
{{- define "bastion.metricsAdditionalLabels" -}}
app.kubernetes.io/component: metrics
{{- end }}
