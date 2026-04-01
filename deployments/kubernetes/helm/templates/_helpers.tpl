{{/*
Expand the name of the chart.
*/}}
{{- define "new-energy-monitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "new-energy-monitoring.fullname" -}}
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
{{- define "new-energy-monitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "new-energy-monitoring.labels" -}}
helm.sh/chart: {{ include "new-energy-monitoring.chart" . }}
{{ include "new-energy-monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "new-energy-monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "new-energy-monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "new-energy-monitoring.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "new-energy-monitoring.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the proper image name
*/}}
{{- define "new-energy-monitoring.image" -}}
{{- $registryName := .Values.global.imageRegistry -}}
{{- $repositoryName := .image.repository -}}
{{- $tag := .image.tag | default .Chart.AppVersion -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for ingress
*/}}
{{- define "new-energy-monitoring.ingress.apiVersion" -}}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.GitVersion -}}
networking.k8s.io/v1
{{- else if semverCompare ">=1.14-0" .Capabilities.KubeVersion.GitVersion -}}
networking.k8s.io/v1beta1
{{- else -}}
extensions/v1beta1
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for HPA
*/}}
{{- define "new-energy-monitoring.hpa.apiVersion" -}}
{{- if semverCompare ">=1.23-0" .Capabilities.KubeVersion.GitVersion -}}
autoscaling/v2
{{- else -}}
autoscaling/v2beta2
{{- end }}
{{- end }}

{{/*
Create annotations for prometheus scraping
*/}}
{{- define "new-energy-monitoring.prometheus.annotations" -}}
prometheus.io/scrape: "true"
prometheus.io/port: "{{ .port }}"
prometheus.io/path: "{{ .path | default "/metrics" }}"
{{- end }}

{{/*
Return the proper Storage Class
*/}}
{{- define "new-energy-monitoring.storageClass" -}}
{{- if .Values.global.storageClass }}
{{- .Values.global.storageClass -}}
{{- else if .Values.persistence.storageClass }}
{{- .Values.persistence.storageClass -}}
{{- else -}}
""
{{- end }}
{{- end }}

{{/*
Renders a value that contains template.
Usage: {{ include "new-energy-monitoring.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "new-energy-monitoring.render" -}}
{{- if typeIs "string" .value }}
{{- tpl .value .context }}
{{- else }}
{{- tpl (.value | toYaml) .context }}
{{- end }}
{{- end }}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "new-energy-monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "new-energy-monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "new-energy-monitoring.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Check if the database is enabled
*/}}
{{- define "new-energy-monitoring.database.enabled" -}}
{{- if .Values.postgresql.enabled }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Check if redis is enabled
*/}}
{{- define "new-energy-monitoring.redis.enabled" -}}
{{- if .Values.redis.enabled }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Check if kafka is enabled
*/}}
{{- define "new-energy-monitoring.kafka.enabled" -}}
{{- if .Values.kafka.enabled }}
true
{{- else }}
false
{{- end }}
{{- end }}

{{/*
Return the database host
*/}}
{{- define "new-energy-monitoring.database.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" .Release.Name -}}
{{- else }}
{{- .Values.externalDatabase.host -}}
{{- end }}
{{- end }}

{{/*
Return the redis host
*/}}
{{- define "new-energy-monitoring.redis.host" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master" .Release.Name -}}
{{- else }}
{{- .Values.externalRedis.host -}}
{{- end }}
{{- end }}

{{/*
Return the kafka brokers
*/}}
{{- define "new-energy-monitoring.kafka.brokers" -}}
{{- if .Values.kafka.enabled }}
{{- printf "%s-kafka:9092" .Release.Name -}}
{{- else }}
{{- .Values.externalKafka.brokers -}}
{{- end }}
{{- end }}
