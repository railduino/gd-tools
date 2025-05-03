#!/bin/bash

# gd-tools Benutzer anlegen, falls er nicht existiert
id -u gd-tools >/dev/null 2>&1 || useradd -r -m -s /bin/bash gd-tools

# Zur docker-Gruppe hinzufügen, falls nicht schon drin
getent group docker | grep -q '\b\(gd-tools\)\b' || usermod -aG docker gd-tools

# Verzeichnisse anlegen und Rechte setzen
{{- range .Dirs }}
install -o gd-tools -g gd-tools -m 755 -d {{ . }}
{{- end }}

echo "prod" >/etc/gd-tools-env
chmod 444 /etc/gd-tools-env
chown root:root /etc/gd-tools-env

