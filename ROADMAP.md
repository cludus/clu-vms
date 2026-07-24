# Flatcar Support Roadmap

## Objective

Add two new include scripts — `gen-flatcar` and `run-flatcar` — that follow
the same architecture as `gen-cloud-init` and `run-cloud-init`, enabling
Flatcar Container Linux VMs managed through the same `cvms` entrypoint.

---

## Phase 0 — Foundation: Shared Helpers & Flatcar Image

**Goal:** Refactor hardcoded helpers into `bin/inc/utils` so they can be
reused by both cloud-init and Flatcar generators. Also ensure the Flatcar
base image is available for later phases.

**Why first:** Every subsequent phase depends on these shared primitives
and on having a Flatcar image to test against.

### Changes

| File | Action |
|---|---|
| `bin/inc/utils` | Add `copy_base_os_image` (generic, accepts `os_name` as 1st arg) |
| `bin/inc/utils` | Add `format_data_disk` (extracted from gen-cloud-init) |
| `bin/inc/gen-cloud-init` | Remove `copy_base_os_image` and `format_data_disk`; call the shared versions from `utils` instead |
| `bin/inc/download` | Add Flatcar image download to `.assets/imgs/flatcar.img` |

### Decisions to finalize in this phase

- **Data disk for Flatcar:** We skip `format_data_disk` for Flatcar hosts —
  just create the raw qcow2 and let the Butane config handle formatting.
- **OS image copy:** `copy_base_os_image` becomes `copy_base_os_image <os_name> <vm_name>`.

### Acceptance tests

1. Run `./bin/cvms download` — verify `flatcar.img` appears in `.assets/imgs/`.
2. Run `./bin/compile` (if templates were not compiled yet), then run
   `./bin/cvms gen-cloud-init` in `test/` — verify it still works
   (image copy, seed ISO, data disk format all succeed).
3. Run `bash -n bin/inc/utils bin/inc/gen-cloud-init bin/inc/download`.

---

## Phase 1 — Template & Compile Infrastructure

**Goal:** Create the Flatcar Butane Jinja2 template and extend `./bin/compile`
to embed it alongside the existing cloud-init templates.

**Why second:** `gen-flatcar` (Phase 3) needs the pre-compiled template
variable to be available at runtime, just like `gen-cloud-init` does.

### Changes

| File | Action |
|---|---|
| `tpls/flatcar_butane.j2` | Create Jinja2 template that renders a Butane YAML skeleton |
| `bin/compile` (`do_compile`) | Read `flatcar_butane.j2` and emit `CVMS_TEMPLATE_FLATCAR_BUTANE` into `bin/inc/templates` |

Note: `build_dist_cvms` wiring for `gen-flatcar` and `run-flatcar` is deferred
to Phase 6, after those scripts exist.

### Template variables (flatcar_butane.j2)

| Variable | Source |
|---|---|
| `{{ name }}` | `host.name` |
| `{{ ip }}` | `host.ip` |
| `{{ gateway }}` | derived from `host.ip` (replace last octet with `.1`) |
| `{{ dns_servers }}` | hardcoded `8.8.8.8, 1.1.1.1` (or from `vms.yml` defaults) |
| `{{ ssh_authorized_keys }}` | `defaults.ssh_authorized_keys` or per-host override (added in Phase 2 — use empty list as placeholder for now) |
| `{{ data_disk_device }}` | `/dev/vdb` |
| `{{ data_disk_mount }}` | `/var/lib/docker` (hardcoded for now) |

### Acceptance tests

1. Run `./bin/compile` — verify it succeeds and `bin/inc/templates` contains
   `CVMS_TEMPLATE_FLATCAR_BUTANE`.
2. Run `bash -n bin/inc/templates bin/compile dist/cvms`.
3. Verify `bin/inc/templates` contains a `CVMS_TEMPLATE_FLATCAR_BUTANE=` line.

---

## Phase 2 — VMS Config Schema Extensions

**Goal:** Add the new `vms.yml` keys that Flatcar needs: SSH authorized keys
and optional SMP/CPU count. Keep backwards compatibility with existing
cloud-init configs.

**Why third:** Phases 3–4 need these values at generation and runtime.
Doing the schema change in isolation avoids mixing config parsing with
script logic.

### Changes

| File | Action |
|---|---|
| `test/vms.yml` | Add `ssh_authorized_keys` to `defaults` and `cpus` to `defaults` |
| `README.md` | Document new keys in the configuration reference |

### New/updated `vms.yml` keys

```yaml
defaults:
  user: alpine
  pass: changeme
  os: alpine
  bridge: vmbr0
  memory: 2G
  storage_size: 100G
  cpus: 2                  # NEW — number of vCPUs (default: 2)
  ssh_authorized_keys:     # NEW — list of SSH public keys
    - "ssh-rsa AAAAB3..."
```

Per-host overrides for `cpus` and `ssh_authorized_keys` are supported
(same pattern as existing `bridge`, `memory`, `vlan` overrides).

### Decisions

- `ssh_authorized_keys` is a YAML list. The Butane template joins elements
  into the Butane `passwd.users[].ssh_authorized_keys` list.
- `cpus` maps to QEMU `-smp $cpus`.
- Flatcar ignores `pass` — it uses SSH keys only. Cloud-init continues to
  use `pass` as before.

### Acceptance tests

1. `yq '.defaults.ssh_authorized_keys[0]' test/vms.yml` returns a key.
2. `yq '.defaults.cpus' test/vms.yml` returns `2`.
3. Existing `./bin/cvms gen-cloud-init` still works with the updated `vms.yml`
   (backwards compatibility verified).

---

## Phase 3 — `gen-flatcar` Script

**Goal:** Create the generation script that produces per-host Flatcar assets:
OS image copy, Butane YAML → Ignition JSON compilation, and data disk creation.

### Changes

| File | Action |
|---|---|
| `bin/inc/gen-flatcar` | Create with `gen-flatcar_help`, `do_gen-flatcar`, `generate_flatcar_host` |
| `bin/cvms` | Source `gen-flatcar`, add help entry, add `gen-flatcar` command dispatch, and add template-variable guard (same pattern as `gen-cloud-init`) |

### `do_gen-flatcar` flow

1. Validate `vms.yml` exists.
2. Validate tools: `yq`, `butane`, `qemu-img`.
3. Validate `.assets/imgs/flatcar.img` exists.
4. Read defaults from `vms.yml`.
5. For each host where `os == "flatcar"` (or `defaults.os == "flatcar"`):
   - Copy `.assets/imgs/flatcar.img` → `.assets/<host>/flatcar.img` (via shared `copy_base_os_image`).
   - Render Jinja2 template → `.assets/<host>/config/flatcar.yaml`.
   - Compile with `butane --pretty --strict` → `.assets/<host>/config/flatcar.ign`.
   - Create `<host>-data.qcow2` if missing (raw, **no** `format_data_disk` — Butane handles formatting).

### Acceptance tests

1. Create a test `vms.yml` with a flatcar host. Run `./bin/cvms gen-flatcar`.
2. Verify `.assets/<host>/flatcar.img` is a copy of `.assets/imgs/flatcar.img`.
3. Verify `.assets/<host>/config/flatcar.yaml` is rendered with correct values.
4. Verify `.assets/<host>/config/flatcar.ign` is valid JSON (Butane output).
5. Verify `<host>-data.qcow2` exists with correct size.
6. Run `bash -n bin/inc/gen-flatcar bin/cvms`.
7. Run `./bin/cvms gen-cloud-init` — verify no regression.

---

## Phase 4 — `run-flatcar` Script

**Goal:** Create the runtime script that boots Flatcar VMs with QEMU,
injecting the Ignition config via `-fw_cfg`.

### Changes

| File | Action |
|---|---|
| `bin/inc/run-flatcar` | Create with `run-flatcar_help`, `do_run-flatcar`, `flatcar_process_host` |
| `bin/cvms` | Source `run-flatcar`, add help entry, add `run-flatcar` command dispatch |

### `do_run-flatcar` flow

1. Parse `--foreground` flag.
2. Validate `vms.yml` exists.
3. Validate tools: `yq`, `qemu-system-x86_64`.
4. Read defaults from `vms.yml`.
5. For each host where `os == "flatcar"` (or `defaults.os == "flatcar"`):
   - Validate artifacts: `.assets/<host>/flatcar.img`, `.assets/<host>/config/flatcar.ign`, `<host>-data.qcow2`.
   - Resolve bridge and memory overrides.
   - Build QEMU command with:
     - `-enable-kvm -cpu host -m $memory -smp $cpus`
     - OS drive, data drive
     - `-netdev bridge,id=net0,br=$bridge`
     - `-device virtio-net-pci,netdev=net0`
     - `-fw_cfg name=opt/org.flatcar-linux/config,file=$ignition`
   - Daemonize with pidfile unless `--foreground`.

### Acceptance tests

1. Run `./bin/cvms run-flatcar --foreground` with a previously generated
   flatcar host (from Phase 3).
2. Verify QEMU starts and the `-fw_cfg` argument appears in `ps aux | grep qemu`.
3. Verify VM boots and is reachable via SSH with the configured key.
4. Run `./bin/cvms stop <host>` — verify graceful shutdown works.
5. Run `bash -n bin/inc/run-flatcar bin/cvms`.
6. Run `./bin/cvms run-cloud-init --foreground` — verify no regression.

---

## Phase 5 — OS-Based Host Filtering

**Goal:** Both cloud-init and Flatcar generators/ runners should only process
hosts matching their OS type. Currently `gen-cloud-init` processes ALL hosts
unconditionally — this needs fixing for correctness in mixed-OS inventories.

### Changes

| File | Action |
|---|---|
| `bin/inc/gen-cloud-init` | In `do_gen-cloud-init`, skip hosts where `os` is not alpine/ubuntu |
| `bin/inc/run-cloud-init` | In `do_run-cloud-init`, skip hosts where `os` is not alpine/ubuntu |
| `bin/inc/gen-flatcar` | Already filters by `os == "flatcar"` (implemented in Phase 3) — verify correctness |
| `bin/inc/run-flatcar` | Already filters by `os == "flatcar"` (implemented in Phase 4) — verify correctness |
| `bin/inc/utils` | (Optional) Add shared helper `host_matches_os <host_json> <expected_os>` |

### Decision

- `os` resolution: use `host.os` if set, otherwise fall back to `defaults.os`.
- A host with `os: flatcar` is processed only by `gen-flatcar`/`run-flatcar`.
- A host with `os: alpine` (or no `os` key + `defaults.os: alpine`) is
  processed only by `gen-cloud-init`/`run-cloud-init`.

### Acceptance tests

1. Create a `vms.yml` with mixed hosts (one alpine, one flatcar).
2. Run `./bin/cvms gen-cloud-init` — verify only the alpine host is processed.
3. Run `./bin/cvms gen-flatcar` — verify only the flatcar host is processed.
4. Run `./bin/cvms run-cloud-init --foreground` — verify only alpine host boots.
5. Run `./bin/cvms run-flatcar --foreground` — verify only flatcar host boots.
6. Run `./bin/cvms stop` — verify both hosts shut down.

---

## Phase 6 — VLAN, Polish & Dist Build

**Goal:** Add VLAN support to the Flatcar pipeline, finalize the distributable
build, and run the full validation suite.

### Changes

| File | Action |
|---|---|
| `tpls/flatcar_butane.j2` | Add `VLAN={{ vlan }}` to the systemd-networkd unit when `vlan` is set |
| `bin/inc/run-flatcar` | Add `vlan=$vlan` to `-netdev` QEMU argument when `vlan` is set |
| `bin/compile` | Add `gen-flatcar` and `run-flatcar` to `build_dist_cvms` required_files and embedded-script sequence |
| `bin/compile` | Update `build_dist_cvms` grep filter to strip source lines for `gen-flatcar` and `run-flatcar` |
| `README.md` | Add Flatcar workflow section |

### Acceptance tests

1. Add `vlan: 10` to a flatcar host. Run `gen-flatcar` — verify VLAN appears
   in the rendered `.network` unit and the Ignition JSON.
2. Run `run-flatcar` — verify `vlan=10` in the QEMU process args.
3. Run `./bin/compile` — verify `dist/cvms` is generated and bash-syntax clean.
4. Run `dist/cvms help` — verify `gen-flatcar` and `run-flatcar` appear.
5. Full regression: run all cloud-init commands and verify no breakage.

---

## Summary: Phase Execution Order

```
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6
  │           │           │           │           │           │           │
  helpers     template    vms.yml    gen-flatcar  run-flatcar  OS filter   VLAN
  download    compile     schema      script       script       fix both    polish
  refactor    infra       extend      wired        wired        gens       dist
```

**Rule:** Do not start phase N until all acceptance tests for phase N−1 pass.
Each phase is self-contained and produces a working (if incomplete) system.
