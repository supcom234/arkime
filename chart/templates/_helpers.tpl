{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "arkime.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "suricata.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

# {{- define "createEsNodeList" -}}
# {{- $local := dict "first" true -}}
# {{- range $k, $v := $.Values.elastic_coordinating_nodes -}}{{- if not $local.first -}},{{- end -}}https://{{- $.Values.es_user -}}:{{- $.Values.es_password -}}@{{- $v | replace "https://" "" -}}{{- $_ := set $local "first" false -}}{{- end -}}
# {{- end -}}
