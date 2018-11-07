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
    {{- if .Values.spectrumConnect.connectionInfo.existingSecret -}}
        {{- .Vaules.spectrumConnect.connectionInfo.existingSecret -}}
    {{- else -}}
        {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-scbe 
    {{- end -}}                                                                                                        
{{- end -}}

{{/*                                                                                                                                                                         
Create the name for ubiquity-db secret                                                                                                                                                 
*/}}                                                                                                                                                                         
{{- define "ibm_storage_enabler_for_containers.ubiquityDbCredentials" -}}                  
    {{- if .Values.genericConfig.ubiquityDbCredentials.existingSecret -}}
        {{- .Vaules.genericConfig.ubiquityDbCredentials.existingSecret -}}
    {{- else -}}
        {{- template "ibm_storage_enabler_for_containers.fullname" . -}}-ubiquitydb 
    {{- end -}}                                                                                                        
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityServiceName" -}}
ubiquity
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityServiceUrl" -}}
api/v1/namespaces/{{ .Release.Namespace }}/services
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDaemonsetName" -}}
ubiquity-k8s-flex
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDaemonsetUrl" -}}
apis/extensions/v1beta1/namespaces/{{ .Release.Namespace }}/daemonsets
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbDeploymentName" -}}
ubiquity-db
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbDeploymentUrl" -}}
apis/extensions/v1beta1/namespaces/{{ .Release.Namespace }}/deployments
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbPodUrl" -}}
api/v1/namespaces/{{ .Release.Namespace }}/pods
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbPvcName" -}}
ibm-ubiquity-db
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbPvcUrl" -}}
api/v1/namespaces/{{ .Release.Namespace }}/persistentvolumeclaims
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbPvName" -}}
ibm-ubiquity-db
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.ubiquityDbPvUrl" -}}
api/v1/persistentvolumes
{{- end -}}

{{- define "ibm_storage_enabler_for_containers.kubernetesDns" -}}
kubernetes.default
{{- end -}}
