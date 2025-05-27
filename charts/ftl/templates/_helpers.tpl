{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ftl.fullname" -}}
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
Common labels
*/}}
{{- define "ftl.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Values.customLabels -}}
{{ toYaml .Values.customLabels }}
{{- end -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "ftl-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: controller
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-runner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: runner
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-provisioner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: provisioner
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-cron.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: cron
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-http-ingress.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: http-ingress
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-timeline.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: timeline
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-lease.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: lease
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-console.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: console
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-admin.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: admin
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "ftl-schema.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ftl.fullname" . }}
app.kubernetes.io/component: schema
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{/*
Common pod configuration, boilerplate that is not generally used
*/}}
{{- define "ftl.commonPodConfig" -}}
    {{- if .nodeSelector }}
    nodeSelector:
      {{- toYaml .nodeSelector | nindent 8 }}
    {{- end }}
    {{- if .affinity }}
    affinity:
      {{- toYaml .affinity | nindent 8 }}
    {{- end }}
    {{- if .topologySpreadConstraints }}
    topologySpreadConstraints:
      {{- toYaml .topologySpreadConstraints | nindent 8 }}
    {{- end }}
    {{- if .tolerations }}
    tolerations:
      {{- toYaml .tolerations | nindent 8 }}
    {{- end }}
{{- end -}}
{{/*
Pod health checks
*/}}
{{- define "ftl.healthProbes" -}}
readinessProbe:
  {{- if .readinessProbe }}
  {{- toYaml .readinessProbe | nindent 12 }}
  {{- else }}
  httpGet:
    path: /healthz
    port: 8892
  initialDelaySeconds: 1
  periodSeconds: 2
  timeoutSeconds: 2
  successThreshold: 1
  failureThreshold: 15
  {{- end }}
{{- end -}}
{{- define "ftl.securityContext" -}}
securityContext:
  runAsNonRoot: true
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
      drop:
          - "ALL"
  seccompProfile:
    type: RuntimeDefault
{{- end -}}
{{- define "ftl.resources" -}}
resources:
  limits:
    cpu: "{{ .resources.limits.cpu }}"
    memory: "{{ .resources.limits.memory }}"
  requests:
    cpu: "{{ .resources.requests.cpu }}"
    memory: "{{ .resources.requests.memory }}"
{{- end -}}
{{- define "ftl-service.environment-variables" -}}
- name: MY_POD_CPU_REQUEST
  valueFrom:
    resourceFieldRef:
      resource: requests.cpu
- name: MY_POD_CPU_LIMIT
  valueFrom:
    resourceFieldRef:
      resource: limits.cpu
- name: MY_POD_MEM_REQUEST
  valueFrom:
    resourceFieldRef:
      resource: limits.memory
- name: MY_POD_MEM_LIMIT
  valueFrom:
    resourceFieldRef:
      resource: limits.memory
{{- end -}}