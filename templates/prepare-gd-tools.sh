#!/bin/bash

# Installiere docker aus dem offiziellen Repo
# https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository

_docker_url="https://download.docker.com/linux/ubuntu"
_docker_asc="/etc/apt/keyrings/docker.asc"
_linux_arch=$(dpkg --print-architecture)
_linux_code=$(source /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
_listd_line="deb [arch=$_linux_arch signed-by=$_docker_asc] $_docker_url $_linux_code stable"

apt-get update
apt-get install ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL $_docker_url/gpg -o $_docker_asc
chmod a+r $_docker_asc
echo "$_listd_line" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin


# Lege den Benutzer gd-tools an und erlaube ihm die Nutzung von Docker

id -u gd-tools >/dev/null 2>&1 || useradd -r -m -s /bin/bash gd-tools
install -o gd-tools -g gd-tools -m 700 -d /home/gd-tools/.ssh
install -o gd-tools -g gd-tools -m 600 /root/.ssh/authorized_keys /home/gd-tools/.ssh

getent group docker | grep -q '\b\(gd-tools\)\b' || usermod -aG docker gd-tools

{{- range .Dirs }}
install -o gd-tools -g gd-tools -m 755 -d {{ . }}
{{- end }}

echo "prod" >/etc/gd-tools-env
chmod 444 /etc/gd-tools-env
chown root:root /etc/gd-tools-env

