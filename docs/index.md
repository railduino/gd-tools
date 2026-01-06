# gd-tools

`gd-tools` is a structured toolkit for building, deploying, and operating
long-running Internet servers and the services running on them,
in a predictable and reproducible way.

It provides a clear lifecycle for systems that evolve over time:
servers that are rebuilt, migrated, or partially replaced —
without losing control or accumulating configuration debt.

---

## Documentation

The primary documentation for gd-tools lives in the **[Wiki](https://github.com/railduino/gd-tools/wiki)**.

The Wiki describes:
- concepts and mental models
- installation and bootstrap steps
- architectural decisions
- operational guidelines

This page serves as an entry point and overview.
It intentionally avoids duplicating detailed documentation.

---

## Tool structure

At its core, gd-tools consists of two main programs that work together:
one used during development and orchestration, and one used on the
production system to apply changes.

On the development system, gd-tools consists of a small set of binaries:

- **gdt**  
  Development and orchestration tool  
  Used to model systems, generate configuration, and prepare deployments.

- **gd-tools**  
  The production executor and agent  
  Responsible for applying changes on the target host.

- **gd-occ**  
  Nextcloud CLI wrapper  
  Used to generate per-instance `occ-*` commands.

- **gd-wp-cli**  
  WordPress CLI wrapper  
  Used to generate per-instance `wp-*` commands.

On the production host, only **gd-tools** itself is installed permanently.

Depending on the configured services, additional per-instance helper
binaries such as `wp-<name>` or `occ-<name>` may be created as copies of
the corresponding development binaries.

No other executables are installed.
Everything else is generated.

---

## Design goals

gd-tools follows a few strict principles:

- reproducible deployments
- minimal mutable configuration
- clear ownership of responsibilities
- no hidden magic
- no snowflake systems

If something can be generated, it is not stored.
If something must be stored, it is explicit.

---

## Typical use cases

gd-tools is well suited for:

- self-hosted infrastructure
- multiple services on a small number of hosts
- long-running servers that evolve over years
- environments where rebuilds must be safe and boring
- operators who want to understand *why* something exists

---

## Getting started

The recommended entry point is:

- **[Installation & Bootstrap](https://github.com/railduino/gd-tools/wiki/01-Installation-&-Bootstrap)**

This chapter establishes the baseline of a gd-tools managed system.
It explains which components are installed, which artefacts are generated,
and which parts are expected to persist over time.

Once this foundation is in place, all further steps build on it.

---

## Core concepts (recommended reading)

These Wiki chapters explain the mental model behind gd-tools.
They are optional, but highly recommended.

- **[Machines vs Services](https://github.com/railduino/gd-tools/wiki/08-Machines-vs-Services)**  
  Why servers and public names are treated as different things,  
  and how this simplifies growth and migration.

- **[Filesystem Layout](https://github.com/railduino/gd-tools/wiki/09-Filesystem-Layout)**  
  How gd-tools uses `/etc`, `/usr/local/bin`, and `/var/gd-tools`,  
  and why backups become unambiguous.

---

## Services and integrations

gd-tools supports a growing set of services and subsystems, including:

- Web services via Apache and PHP-FPM
- CMS installations (e.g. WordPress, MediaWiki)
- Nextcloud and/or ownCloud Infinite Scale
- Mail infrastructure components
- DNS and certificate automation
- Centralized logging
- Deterministic backups

New services are integrated by fitting into the existing lifecycle —
not by adding ad-hoc scripts.

---

## Repository layout

This repository contains:

- source code for all binaries
- embedded templates for generated configuration
- baseline assets used during bootstrap

The repository itself is part of the system’s reproducibility.

---

## Philosophy

gd-tools encodes operational experience accumulated over many years.

Instead of hiding complexity, it aims to:
- make it explicit
- make it inspectable
- make it repeatable

The goal is not speed.
The goal is stability.

