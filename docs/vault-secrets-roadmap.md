# Vault Secrets — Implementation Roadmap

A GPG-backed secrets vault for `clu-vms`. Two implementations of
[`VaultSecrets`](../internal/spec/secrets.go) will be delivered:

1. **Pure Go** — [`github.com/ProtonMail/go-crypto`](https://github.com/ProtonMail/go-crypto) (no CGO, default).
2. **GPGME backend** — [`github.com/proglottis/gpgme`](https://github.com/proglottis/gpgme) (CGO, system `gpg-agent`).

## Design principles

- **Public keys only on disk.** Private keys live in the user's keyring/agent and are loaded per-session.
- **At least one recipient at all times.** `RemoveKey` on the last key is a hard error.
- **Multi-recipient messages.** A secret is a single OpenPGP message encrypted to all current recipients. OpenPGP handles per-recipient subkeys natively.
- **Re-encryption on roster changes.** `AddKey` / `RemoveKey` rewrites every existing `*.gpg` against the new recipient set.
- **Two implementations, one contract.** Both backends must satisfy `VaultSecrets` byte-for-byte and pass the same test suite.

## Interface contract (target)

```go
type VaultSecrets interface {
    Init(path string) error

    UnlockPrivateKey(armoredKey string, passphrase []byte) error
    SetPassphraseFunc(fn func(fingerprint string) ([]byte, error))

    GetSecret(name string) (string, error)
    SetSecret(name, value string) error
    ListSecrets() ([]string, error)
    ListSecretsWithPrefix(prefix string) ([]string, error)

    ListKeys() ([]string, error)            // fingerprints
    AddKey(armoredPublicKey string) error  // ASCII-armored public key
    RemoveKey(fingerprint string) error
}
```

> Rationale for changes vs the current draft:
> - `Init(path)` pins the vault root.
> - `UnlockPrivateKey` / `SetPassphraseFunc` are mandatory for `GetSecret` since private keys are never persisted.
> - `AddKey` takes an armored block (single contract); the gpgme backend imports it before storing the fingerprint.
> - `RemoveKey` takes a fingerprint (string), not an armored block.

## Storage layout

```
<vaultDir>/
  .keys.json          # {"recipients": ["<fpr1>", "<fpr2>", ...]}
  keys/<fpr>.asc      # ASCII-armored public key per recipient
  secrets/<name>.gpg  # one armored OpenPGP message per secret
```

- `Init` creates the directory tree if missing.
- `Init` refuses to operate if `<vaultDir>/.keys.json` is malformed.
- Empty recipient set is **not allowed** — first run must seed at least one key.

## Phase 0 — Interface lock

- [x] Agree on the refined `VaultSecrets` interface.
- [ ] Add interface tests (in-memory mock) used by both backends.

## Phase 1 — Pure-Go implementation (`internal/impl/secrets/protonmail/`)

Uses `github.com/ProtonMail/go-crypto/openpgp`.

### 1.1 Skeleton & `Init`
- Implement `Init(path)`: create `keys/`, `secrets/`, and seed `.keys.json` from existing `.asc` files if present.
- Validate that every fingerprint in `.keys.json` has a corresponding `.asc`.

### 1.2 Recipient registry
- In-memory cache of `*openpgp.Entity` (public half) keyed by fingerprint.
- `ListKeys` returns fingerprints from `.keys.json`.

### 1.3 `AddKey(armoredPublicKey)`
- Parse armored block into `*openpgp.Entity`.
- Extract primary fingerprint; reject duplicates.
- Persist `keys/<fpr>.asc` and update `.keys.json`.
- Walk every `secrets/*.gpg`, decrypt with current in-memory private keys (if unlocked), re-encrypt to the **new** recipient set, write back.

### 1.4 `RemoveKey(fingerprint)`
- Refuse if it is the last fingerprint in `.keys.json`.
- Drop entry, delete `keys/<fpr>.asc`.
- Re-encrypt every `secrets/*.gpg` to the remaining recipients.

### 1.5 `SetSecret` / `GetSecret`
- `SetSecret`: serialize plaintext, `openpgp.Encrypt` to all recipients, write `secrets/<name>.gpg`.
- `GetSecret`: read file, look up a recipient whose private key is unlocked, call `openpgp.Decrypt` with the entity list and a passphrase callback (`SetPassphraseFunc`).
- Refuse `SetSecret` if zero recipients are configured.

### 1.6 Decryption bootstrap
- `UnlockPrivateKey(armoredKey, passphrase)`: read armored private key, unlock with passphrase, keep the entity in memory (never written to disk).
- `SetPassphraseFunc`: store callback for lazy prompting.

### 1.7 Listing
- `ListSecrets`: read directory, strip `.gpg`.
- `ListSecretsWithPrefix`: filter `ListSecrets` by prefix.

### 1.8 Tests
- Generate `openpgp.Entity` pairs in-test (no real GPG needed).
- Cover: round-trip, multi-recipient decrypt, `AddKey` re-encrypt, `RemoveKey` last-key error, prefix listing, passphrase callback invocation.

## Phase 2 — GPGME implementation (`internal/impl/secrets/gpgme/`)

Uses `github.com/proglottis/gpgme`. **CGO enabled; gated by a build tag.**

### 2.1 Build setup
- Add `//go:build gpgme` tag (or use a separate package directory).
- Document the C dependency (`libgpgme-dev` on Debian/Ubuntu, `gpgme` on macOS via Homebrew).
- Add a `mise.toml` task to install the C lib for dev environments.

### 2.2 Skeleton & `Init`
- `Init(path)`: same directory layout; populate in-memory fingerprint cache from `keys/*.asc` (we still mirror public keys to disk for parity and portability).
- Open a `gpgme.Context` in `C_NORMAL` mode.

### 2.3 Recipient registry
- `ListKeys`: enumerate fingerprints from `.keys.json`.

### 2.4 `AddKey(armoredPublicKey)`
- `gpgme.Context.Import` the armored block.
- Persist `keys/<fpr>.asc` and update `.keys.json`.
- Re-encrypt every `secrets/*.gpg` to the new recipient set using the in-process keyring.

### 2.5 `RemoveKey(fingerprint)`
- Refuse if last recipient.
- `gpgme.Context.KeyDelete` for that fingerprint.
- Drop from `.keys.json`, delete `keys/<fpr>.asc`.
- Re-encrypt every `secrets/*.gpg` to remaining recipients.

### 2.6 `SetSecret` / `GetSecret`
- `SetSecret`: `gpgme.Encrypt` to all recipients → `secrets/<name>.gpg`.
- `GetSecret`: `gpgme.Decrypt` — passphrase is handled by `gpgme`'s pinentry loop (no in-process callback needed). `SetPassphraseFunc` can be ignored or used to pre-supply via `gpgme.SetPinentryMode` only if the user opts in.

### 2.7 Decryption bootstrap
- `UnlockPrivateKey`: no-op for gpgme (private keys live in `gpg-agent`). Method exists only to satisfy the interface.
- `SetPassphraseFunc`: optional override of the agent pinentry; left as a future hook.

### 2.8 Tests
- Tag-gated integration tests using `gpgme`'s in-process engine where available, or a temp `GNUPGHOME`.
- Same coverage matrix as Phase 1.

## Phase 3 — Factory & selection

- [ ] `internal/impl/secrets/factory.go`: select backend by:
  1. Build tag (compile-time).
  2. Config flag (`vault.backend: protonmail | gpgme`) for runtime swap when both are compiled in.
- [ ] Wire into CLI subcommands under `cmd/cvms/`.

## Phase 4 — Hardening

- [ ] Atomic writes (write to `*.tmp`, `fsync`, rename) for `.gpg` and `.keys.json`.
- [ ] File mode `0600` on every write.
- [ ] Lock file (`flock` on `.keys.json`) to serialize concurrent CLI invocations.
- [ ] Backup/export: dump `keys/*.asc` + `secrets/*.gpg` to a tarball.
- [ ] Document the **OpenPGP revocation caveat**: `RemoveKey` stops future encryption to that recipient but does not retroactively revoke; historical messages remain decryptable by anyone who still holds the private key.

## Phase 5 — Documentation

- [ ] User guide in `docs/vault-usage.md`: init, add first key, store/retrieve secrets, rotate keys.
- [ ] Threat model section: who can read secrets, what protects them, what does not.
- [ ] Migration notes if a non-GPG vault already exists.

---

## Milestones

| # | Deliverable | Exit criteria |
|---|---|---|
| M1 | Interface finalized + mock test suite | Both backends can be written against the same contract |
| M2 | Pure-Go backend (Phase 1) | All Phase 1 tests green, CLI usable without `gpg-agent` |
| M3 | GPGME backend (Phase 2) | Same test matrix passes under `-tags gpgme` |
| M4 | Factory + CLI wiring | `cvms secrets` subcommands work with either backend |
| M5 | Hardening + docs | Atomic writes, lockfile, user guide, threat model published |
