# gd-tools

**A pragmatic Go toolchain for self-hosted infrastructure.**

gd-tools is a set of composable tools for building, provisioning, and operating
self-hosted systems in a reproducible and transparent way.

Instead of relying on large, generic frameworks, gd-tools focuses on explicit
models, typed configuration, and deterministic workflows. Systems are described
once, generated from source, and deployed in a controlled and repeatable manner.

gd-tools is designed for operators who want to understand and own their stack —
from initial setup to long-term maintenance — without hiding complexity behind
layers of abstraction.

It favors clarity over convenience, reproducibility over flexibility, and code
over ad-hoc scripting. All artifacts are generated, not edited by hand. All
changes are intentional, reviewable, and repeatable.

gd-tools is opinionated by design and aims to stay small, inspectable, and
adaptable over time.

While gd-tools is designed to stay adaptable, it does make a set of explicit
assumptions about the target environment.

The reference environment for gd-tools is a self-hosted setup on Hetzner
infrastructure. This includes virtual machines hosted as Hetzner Cloud servers,
DNS management via the Hetzner Cloud API, and off-site backups using Hetzner
Storage Boxes.

These assumptions are not hard requirements, but they strongly influence the
default workflows, integrations, and abstractions provided by gd-tools. Other
providers can be integrated (e.g. the IONOS DNS API), but Hetzner is treated
as the primary, well-tested baseline.

---

## Documentation

Comprehensive documentation lives in the project wiki:

👉 **https://github.com/railduino/gd-tools/wiki**

Start here:
- [Installation & Bootstrap](https://github.com/railduino/gd-tools/wiki/01-Installation-&-Bootstrap)
- [Your first production server](https://github.com/railduino/gd-tools/wiki/02-Your-first-production-server)

Further reading:
- [Design Goals](https://github.com/railduino/gd-tools/wiki/90-Design-Goals)
- [Non-Goals](https://github.com/railduino/gd-tools/wiki/91-Non-Goals)
- [Core Concepts](https://github.com/railduino/gd-tools/wiki/92-Core-Concepts)
- [Reference Environment: Hetzner](https://github.com/railduino/gd-tools/wiki/93-Reference-Environment:-Hetzner)

