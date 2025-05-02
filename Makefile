# Copyright (c) 2025 Volker Wiegand

VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +"%Y-%m-%d/%H:%M:%S")
TEMPLATES := $(shell find templates/ -type f)
LOCALES := $(shell find locales/ -type f)

all: replaces gd-tools

gd-tools: *.go messages.pot $(TEMPLATES) $(LOCALES)
	go mod tidy
	go fmt ./...
	@xgettext --language=Python -kT -kT_out -kT_err  -kTf_out -kTf -kTn:1,2 -kTnf:1,2 -o messages.pot *.go
	@msgmerge --update locales/de_DE/LC_MESSAGES/messages.po messages.pot
	@msgattrib --clear-fuzzy -o locales/de_DE/LC_MESSAGES/messages.po locales/de_DE/LC_MESSAGES/messages.po
	@msgmerge --update locales/en_US/LC_MESSAGES/messages.po messages.pot
	@msgattrib --clear-fuzzy -o locales/en_US/LC_MESSAGES/messages.po locales/en_US/LC_MESSAGES/messages.po
	go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o gd-tools
	sudo install gd-tools /usr/local/bin

de:
	vim locales/de_DE/LC_MESSAGES/messages.po

en:
	vim locales/en_US/LC_MESSAGES/messages.po

translate:
	@msgmerge --update locales/de_DE/LC_MESSAGES/messages.po messages.pot
	@msgattrib --clear-fuzzy -o locales/de_DE/LC_MESSAGES/messages.po locales/de_DE/LC_MESSAGES/messages.po
	@msgmerge --update locales/en_US/LC_MESSAGES/messages.po messages.pot
	@msgattrib --clear-fuzzy -o locales/en_US/LC_MESSAGES/messages.po locales/en_US/LC_MESSAGES/messages.po

clean:
	rm -f gd-tools

replaces:
	@grep -q 'replace github.com/docker/docker' go.mod || \
		go mod edit -replace=github.com/docker/docker=github.com/docker/docker@v24.0.7+incompatible
	@grep -q 'replace github.com/docker/distribution' go.mod || \
		go mod edit -replace=github.com/docker/distribution=github.com/docker/distribution@v2.7.1+incompatible

pull:
	git fetch
	git merge

push:
	git add .
	git commit -a
	-git push

init:
	go mod init github.com/railduino/gd-tools

zip:
	tar -c -v -z -f /tmp/gd-tools.tgz Makefile *.go templates locales

