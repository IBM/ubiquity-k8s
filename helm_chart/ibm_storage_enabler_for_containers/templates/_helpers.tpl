{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ibm_storage_enabler_for_containers.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ibm_storage_enabler_for_containers.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ibm_storage_enabler_for_containers.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name for the scbe secret
*/}}
{{- define "ibm_storage_enabler_for_containers.scbeCredentials" -}}
    {{- if .Values.ubiquity.spectrumConnect -}}
        {{- if .Values.ubiquity.spectrumConnect.connectionInfo.existingSecret -}}
            {{- .Values.ubiquity.spectrumConnect.connectionInfo.existingSecret -}}
        {{- else -}}
            {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-scbe 
        {{- end -}}
    {{- else -}}
        {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-scbe 
    {{- end -}}
{{- end -}}

{{/*
Create the name for the spectrum scale secret
*/}}
{{- define "ibm_storage_enabler_for_containers.spectrumscaleCredentials" -}}
    {{- if .Values.ubiquity.spectrumScale -}}
        {{- if .Values.ubiquity.spectrumScale.connectionInfo.existingSecret -}}
            {{- .Values.ubiquity.spectrumScale.connectionInfo.existingSecret -}}
        {{- else -}}
            {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-scale
        {{- end -}}
	{{- else -}}
        {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-scale
    {{- end -}}
{{- end -}}

{{/*
Create the name for ubiquity-db secret
*/}}
{{- define "ibm_storage_enabler_for_containers.ubiquityDbCredentials" -}}
    {{- if .Values.ubiquityDb.dbCredentials.existingSecret -}}
        {{- .Values.ubiquityDb.dbCredentials.existingSecret -}}
    {{- else -}}
        {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-ubiquitydb 
    {{- end -}}
{{- end -}}

{{/*
Create the name of storageClass for ubiquity-db pvc
*/}}
{{- define "ibm_storage_enabler_for_containers.ubiquityDbStorageClass" -}}
    {{- if .Values.ubiquityDb.persistence.storageClass.existingStorageClass -}}
        {{- .Values.ubiquityDb.persistence.storageClass.existingStorageClass -}}
    {{- else -}}
        {{ .Values.ubiquityDb.persistence.storageClass.storageClassName }}
    {{- end -}}
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.helmLabels" -}}
app: {{ template "ibm_storage_enabler_for_containers.name" . }}
chart: {{ template "ibm_storage_enabler_for_containers.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
{{- end -}}
