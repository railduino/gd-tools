# gd-tools

gd-tools is a structured toolkit for building, deploying,
and operating self-hosted services in a predictable and reproducible way.

It provides a clear lifecycle for long-lived systems:
systems that grow over time, are rebuilt, migrated, or partially replaced —
without losing control or accumulating configuration debt.

---

## Documentation

The primary documentation for gd-tools lives in the **Wiki**.

The Wiki describes:
- concepts and mental models
- installation and bootstrap steps
- architectural decisions
- operational guidelines

This page serves as an entry point and overview.
It intentionally avoids duplicating detailed documentation.

---

## What gd-tools is

gd-tools is a set of cooperating tools with clearly separated responsibilities:

- **gdt**  
  Development and orchestration tool  
  Used to model systems, generate configuration, and prepare deployments.
  Installed only on the **Development Workstation**.

- **gd-tools**  
  Minimal production-side executor (**agent**)  
  Installed on each **Production Host**.
  Applies changes deterministically on the target host.

More or less, anything else is generated as part of the provisioning workflow.

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

### → **Your first host**

This introduces:
- the base directory structure
- the global baseline configuration via `gdt basics`
- the separation between generated artefacts and persistent data

From there, services can be added incrementally.

---

## Core concepts (recommended reading)

These Wiki chapters explain the mental model behind gd-tools.
They are optional, but highly recommended.

- **08 Machines vs Services**  
  Why servers and public names are treated as different things,  
  and how this simplifies growth and migration.

- **09 Filesystem Layout**  
  How gd-tools uses `/etc`, `/usr/local/bin`, and `/var/gd-tools`,  
  and why backups become unambiguous.

---

## Services and integrations

gd-tools supports a growing set of services and subsystems, including:

- Web services via Apache and PHP-FPM
- CMS installations (e.g. WordPress, MediaWiki)
- Nextcloud
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
- documentation under `/docs`

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

---

## Next steps

- Start with **Your first host** in the Wiki
- Read **08 Machines vs Services**
- Review **09 Filesystem Layout**
- Then model your first service

From there, gd-tools stays out of your way.

