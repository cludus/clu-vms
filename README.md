# CVMS — Cludus Cloud Virtual Machine Setup

Bootstraps cloud-init based VMs using stock cloud images.

## Quick Start

```bash
./bin/cvms help
```

The `bin/cvms` script is the **main entrypoint** for all operations.

## Directory Layout

```
bin/cvms          Main entrypoint
bin/inc/          Include scripts (sourced by cvms)
tpls/             Jinja2 cloud-init templates
imgs/             Downloaded cloud images
```
