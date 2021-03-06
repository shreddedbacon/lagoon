{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "mongodb-single.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "mongodb-single.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mongodb-single.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create full hostname for autogenerated hosts
*/}}
{{- define "mongodb-single.autogeneratedHost" -}}
{{- printf "%s.%s" .Release.Name .Values.routesAutogenerateSuffix | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "mongodb-single.labels" -}}
helm.sh/chart: {{ include "mongodb-single.chart" . }}
{{ include "mongodb-single.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ include "mongodb-single.lagoonLabels" . }}
{{- end -}}

{{/*
Add annotations
*/}}
{{- define "mongodb-single.annotations" -}}
{{ if .Values.annotations }}
{{- toYaml .Values.annotations }}
{{- end }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "mongodb-single.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mongodb-single.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create a PriorityClassName.
(this is based on the Lagoon Environment Type)).
*/}}
{{- define "mongodb-single.lagoonPriority" -}}
{{- printf "lagoon-priority-%s" .Values.environmentType }}
{{- end -}}

{{/*
Lagoon Labels
*/}}
{{- define "mongodb-single.lagoonLabels" -}}
lagoon.sh/service: {{ .Release.Name }}
lagoon.sh/service-type: {{ .Chart.Name }}
lagoon.sh/project: {{ .Values.project }}
lagoon.sh/environment: {{ .Values.environment }}
lagoon.sh/environmentType: {{ .Values.environmentType }}
lagoon.sh/buildType: {{ .Values.buildType }}
{{- end -}}

{{/*
Annotations
*/}}
{{- define "mongodb-single" -}}
lagoon.sh/version: {{ .Values.lagoonVersion | quote }}
{{- if .Values.branch }}
lagoon.sh/branch: {{ .Values.branch | quote }}
{{- end }}
{{- if .Values.prNumber }}
lagoon.sh/prNumber: {{ .Values.prNumber | quote }}
lagoon.sh/prHeadBranch: {{ .Values.prHeadBranch | quote }}
lagoon.sh/prBaseBranch: {{ .Values.prBaseBranch | quote }}
{{- end }}
{{- end -}}
