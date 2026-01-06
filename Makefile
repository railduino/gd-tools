# Copyright (c) 2025-2026 Volker Wiegand

VERSION   := $(shell git describe --tags --always 2>/dev/null || echo "dev")

SRC       := $(shell find . -name '*.go')
ASSETS    := $(shell find assets/ -type f)
TEMPLATES := $(shell find templates/templates/ -type f)

.PHONY: all gdt gd-tools gd-occ gd-wp-cli clean test completion pull push

all: gdt gd-tools gd-occ gd-wp-cli

gdt: bin/gdt

gd-tools: bin/gd-tools

gd-occ: bin/gd-occ

gd-wp-cli: bin/gd-wp-cli

bin/gdt: $(SRC) $(ASSETS) $(TEMPLATES)
	@mkdir -p templates/templates/assets
	cp -a assets/* templates/templates/assets/
	go mod tidy
	go fmt ./...
	go vet ./...
	go build -o bin/gdt \
		-ldflags "-X 'main.version=$(VERSION)'" \
		./cmd/gdt
	sudo install bin/gdt /usr/local/bin

bin/gd-tools: $(SRC) $(TEMPLATES)
	go build -o bin/gd-tools \
		-ldflags "-X 'main.version=$(VERSION)'" \
		./cmd/gd-tools
	sudo install bin/gd-tools /usr/local/bin

bin/gd-occ: $(SRC) $(TEMPLATES)
	go build -o bin/gd-occ \
		-ldflags "-X 'main.version=$(VERSION)'" \
		./cmd/gd-occ
	sudo install bin/gd-occ /usr/local/bin

bin/gd-wp-cli: $(SRC) $(TEMPLATES)
	go build -o bin/gd-wp-cli \
		-ldflags "-X 'main.version=$(VERSION)'" \
		./cmd/gd-wp-cli
	sudo install bin/gd-wp-cli /usr/local/bin

test:
	go test ./...

clean:
	rm -f bin/gdt bin/gd-tools bin/gd-occ bin/gd-wp-cli

completion: bin/gdt
	sudo gdt completion --save

pull:
	git pull --rebase

push:
	git add .
	git commit -a
	git push

