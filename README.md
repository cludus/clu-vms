# CVMS — Cludus Cloud Virtual Machine Setup

Bootstraps cloud-init based VMs using stock cloud images with
QEMU/KVM and bridge networking.

## Quick Start

```bash
./bin/cvms help
```

The `bin/cvms` script is the **main entrypoint** for all operations.

All commands must be run from a directory containing a `vms.yml` file
(see [Configuration](#configuration)).

## Commands

| Command | Description |
|---|---|
| `cvms download` | Download latest Alpine Linux and Ubuntu Server cloud images |
| `cvms gen-cloud-init` | Generate per-VM cloud-init config, seed ISO, OS copy, and storage |
| `cvms run-cloud-init [--foreground]` | Start all VMs (daemonized by default) |
| `cvms stop [host-name]` | Graceful ACPI shutdown of VMs (all, or a single host) |

## Workflow for Cloud Init

```bash
cd test                      # directory with vms.yml
../bin/cvms download         # 1. Download base OS images to .assets/imgs/
../bin/cvms gen-cloud-init   # 2. Generate per-VM assets under .assets/<vm>/
../bin/cvms run-cloud-init   # 3. Start all VMs (daemonized)
../bin/cvms stop             # 4. Gracefully shut down all VMs
```

## Configuration

A `vms.yml` file defines defaults, bridge networking, and hosts:

```yaml
defaults:
  user: alpine          # username
  pass: changeme        # password (optional for flatcar if ssh_authorized_keys is set)
  os: alpine            # default OS for VMs
  bridge: vmbr0         # default bridge
  memory: 2G            # RAM per VM (optional, defaults to 2G)
  storage_size: 100G    # data disk size (optional, defaults to 100G)
  cpus: 2               # number of vCPUs (optional, defaults to 2)
  ssh_authorized_keys:  # optional SSH public keys list
    - "ssh-rsa AAAAB3..."

bridge:
  name: vmbr0           # bridge interface name
  phys_if: eth0         # physical interface to enslave
  ip: 192.168.100.1/24  # bridge IP (should match host's subnet)
  create: true          # auto-create bridge if missing

hosts:
  - name: vm1
    ip: 192.168.100.201
  - name: vm2
    ip: 192.168.100.202
```

Per-host overrides for `bridge`, `memory`, `vlan`, `cpus`, and
`ssh_authorized_keys` are supported.

For Flatcar hosts, at least one authentication method must be configured:
either a non-empty `pass` or at least one `ssh_authorized_keys` entry.

## Assets Directory

Generated and downloaded assets live under `.assets/` in the working
directory (customizable via the `ASSETS_DIR` environment variable):

```
.assets/
├── imgs/              Downloaded base OS images (alpine.img, ubuntu.img)
└── <vm-name>/         Per-VM generated assets
    ├── alpine.img     Copy of base OS image
    ├── config/        Rendered cloud-init files
    │   ├── user-data
    │   ├── meta-data
    │   └── network-config
    └── seed.img       Cloud-init seed ISO
```

Storage disks (`<vm-name>-data.qcow2`) are kept in the working directory
alongside `vms.yml`.

VM runtime files (`qemu.pid`) are kept in `<vm-name>/` under the working
directory.

## Dependencies

- **Runtime:** `qemu-system-x86_64`, `qemu-img`, `qemu-nbd`, `genisoimage`, `yq`
- **Python:** `jinja2` (for cloud-init template rendering)
- **Linux tools:** `ip`, `sudo`, `sfdisk`, `mkfs.ext4`, `nbd`
- **QEMU bridge:** `/etc/qemu/bridge.conf` must allow the bridge (e.g. `allow vmbr0`),
  and `/usr/lib/qemu/qemu-bridge-helper` needs the setuid bit (`chmod u+s`)

## Directory Layout

```
bin/cvms            Main entrypoint
bin/inc/            Include scripts (download, gen-cloud-init, run-cloud-init, stop, utils)
tpls/               Jinja2 cloud-init templates (meta-data, user-data, network-config)
demos/              Example configurations
test/               Test vms.yml configurations
```
