require ["copy", "redirect", "envelope"];

# Forwarding for: {{ .Email }}

if envelope :is "to" "{{ .Email }}" {
{{- if .Dismiss }}
  {{- range .Forwards }}
  redirect "{{ . }}";
  {{- end }}
  stop;
{{- else }}
  {{- range .Forwards }}
  redirect :copy "{{ . }}";
  {{- end }}
{{- end }}
}

