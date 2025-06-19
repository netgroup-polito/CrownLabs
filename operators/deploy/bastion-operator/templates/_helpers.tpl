{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bastion-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bastion-operator.fullname" -}}
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
{{- define "bastion-operator.version" -}}
{{- if .Values.global }}
{{- .Values.image.tag | default .Values.global.version | default .Chart.AppVersion }}
{{- else }}
{{- .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bastion-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bastion-operator.labels" -}}
helm.sh/chart: {{ include "bastion-operator.chart" . }}
{{ include "bastion-operator.selectorLabels" . }}
app.kubernetes.io/version: {{ include "bastion-operator.version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bastion-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bastion-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Metrics selector additional labels
*/}}
{{- define "bastion-operator.metricsAdditionalLabels" -}}
app.kubernetes.io/component: metrics
{{- end }}
