{{- define "erp-system.serviceAccountName" -}}
{{- printf "%s" (include "erp-system.fullname" .) | trunc 63 -}}
{{- end -}}

{{- define "erp-system.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s" .Release.Name | trunc 63 -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 -}}
{{- end -}}
{{- end -}}

{{- define "erp-system.clientCommandService.name" -}}
{{- printf "client-command-service" -}}
{{- end -}}

{{- define "erp-system.authService.name" -}}
{{- printf "auth-service" -}}
{{- end -}}

{{- define "erp-system.namespace" -}}
{{- .Release.Namespace | default "erp-system" -}}
{{- end -}}
