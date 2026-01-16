{{/*
Expand the name of the chart.
*/}}
{{- define "idekube.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "idekube.fullname" -}}
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
{{- define "idekube.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "idekube.labels" -}}
helm.sh/chart: {{ include "idekube.chart" . }}
{{ include "idekube.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "idekube.selectorLabels" -}}
app.kubernetes.io/name: {{ include "idekube.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "idekube.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "idekube.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Controller fullname
*/}}
{{- define "idekube.controller.fullname" -}}
{{- printf "%s-controller" (include "idekube.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Controller labels
*/}}
{{- define "idekube.controller.labels" -}}
{{ include "idekube.labels" . }}
app.kubernetes.io/component: controller
{{- end }}

{{/*
Controller selector labels
*/}}
{{- define "idekube.controller.selectorLabels" -}}
{{ include "idekube.selectorLabels" . }}
app.kubernetes.io/component: controller
{{- end }}

{{/*
Housekeeper fullname
*/}}
{{- define "idekube.housekeeper.fullname" -}}
{{- printf "%s-housekeeper" (include "idekube.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Housekeeper labels
*/}}
{{- define "idekube.housekeeper.labels" -}}
{{ include "idekube.labels" . }}
app.kubernetes.io/component: housekeeper
{{- end }}

{{/*
Housekeeper selector labels
*/}}
{{- define "idekube.housekeeper.selectorLabels" -}}
{{ include "idekube.selectorLabels" . }}
app.kubernetes.io/component: housekeeper
{{- end }}

{{/*
Frontend fullname
*/}}
{{- define "idekube.frontend.fullname" -}}
{{- printf "%s-frontend" (include "idekube.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "idekube.frontend.labels" -}}
{{ include "idekube.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "idekube.frontend.selectorLabels" -}}
{{ include "idekube.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
PostgreSQL fullname
*/}}
{{- define "idekube.postgresql.fullname" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "idekube.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "postgresql" }}
{{- end }}
{{- end }}

{{/*
Redis fullname
*/}}
{{- define "idekube.redis.fullname" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis" (include "idekube.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "redis" }}
{{- end }}
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "idekube.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}
