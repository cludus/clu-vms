package mock

// Package mock provides an in-memory reference implementation of
// spec.VaultSecrets intended for tests and as a behavior reference for
// real backends.
//
// Mock performs NO encryption. Secrets are stored as plaintext with a
// "recipient list" attached. This is acceptable for testing the contract
// only — never use Mock for real secrets.
//
// Mock is not safe for concurrent use; tests are expected to be sequential.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"clu-vms/internal/spec"
)

// Fingerprint derives a stable 40-char uppercase hex fingerprint from an
// armored block. Real OpenPGP fingerprints come from the public key
// packet; mock uses SHA-256 of the armored text so it is reproducible,
// dependency-free, and format-compatible (40 uppercase hex chars).
//
// Use Fingerprint when wiring the conformance suite to Mock. Real
// backends will provide their own fingerprint function (e.g. parsed from
// the OpenPGP packet), but the conformance suite is fingerprint-agnostic:
// it goes through Backend.Fingerprint supplied by the Factory.
func Fingerprint(armored string) string {
	sum := sha256.Sum256([]byte(armored))
	return strings.ToUpper(hex.EncodeToString(sum[:])[:40])
}

// Mock is an in-memory spec.VaultSecrets.
type Mock struct {
	root         string
	initialized  bool
	fingerprints []string
	pubKeys      map[string]string // fpr -> armored public key
	privKeys     map[string]string // fpr -> armored private key (added via UnlockPrivateKey)
	secretMeta   map[string][]string
	secrets      map[string]string
	passphraseFn func(fingerprint string) ([]byte, error)
}

// Compile-time check that Mock satisfies the spec.
var _ spec.VaultSecrets = (*Mock)(nil)

// New returns an uninitialized Mock. Init must be called before use.
func New() *Mock {
	return &Mock{
		pubKeys:    make(map[string]string),
		privKeys:   make(map[string]string),
		secretMeta: make(map[string][]string),
		secrets:    make(map[string]string),
	}
}

func (m *Mock) Init(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", spec.ErrInvalidVault)
	}
	m.root = path
	m.initialized = true
	return nil
}

func (m *Mock) UnlockPrivateKey(armoredKey string, passphrase []byte) error {
	if !m.initialized {
		return spec.ErrInvalidVault
	}
	if armoredKey == "" {
		return fmt.Errorf("%w: empty armored key", spec.ErrInvalidKey)
	}
	if len(passphrase) == 0 {
		return fmt.Errorf("%w: empty passphrase", spec.ErrInvalidKey)
	}
	// Mock does not verify the passphrase against the key. Real backends
	// must; the contract is that UnlockPrivateKey returns nil only if the
	// supplied passphrase actually unlocks the key.
	fpr := Fingerprint(armoredKey)
	m.privKeys[fpr] = armoredKey
	return nil
}

func (m *Mock) SetPassphraseFunc(fn func(fingerprint string) ([]byte, error)) {
	m.passphraseFn = fn
}

func (m *Mock) GetSecret(name string) (string, error) {
	if !m.initialized {
		return "", spec.ErrInvalidVault
	}
	if name == "" {
		return "", fmt.Errorf("%w: empty name", spec.ErrSecretNotFound)
	}
	val, ok := m.secrets[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", spec.ErrSecretNotFound, name)
	}
	recips := m.secretMeta[name]
	if len(recips) == 0 {
		return "", spec.ErrNoRecipients
	}

	// Look for an available private key among the configured recipients.
	for _, fpr := range recips {
		if _, ok := m.privKeys[fpr]; ok {
			return val, nil
		}
	}

	// Try the passphrase callback for any recipient that was not pre-loaded.
	if m.passphraseFn != nil {
		for _, fpr := range recips {
			if _, ok := m.privKeys[fpr]; ok {
				continue
			}
			if _, err := m.passphraseFn(fpr); err == nil {
				m.privKeys[fpr] = "lazy:" + fpr
				return val, nil
			}
		}
	}

	return "", spec.ErrDecryptionFailed
}

func (m *Mock) SetSecret(name, value string) error {
	if !m.initialized {
		return spec.ErrInvalidVault
	}
	if name == "" {
		return fmt.Errorf("%w: empty name", spec.ErrInvalidKey)
	}
	if len(m.fingerprints) == 0 {
		return spec.ErrNoRecipients
	}
	recips := append([]string(nil), m.fingerprints...)
	m.secretMeta[name] = recips
	m.secrets[name] = value
	return nil
}

func (m *Mock) ListSecrets() ([]string, error) {
	if !m.initialized {
		return nil, spec.ErrInvalidVault
	}
	names := make([]string, 0, len(m.secrets))
	for n := range m.secrets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

func (m *Mock) ListSecretsWithPrefix(prefix string) ([]string, error) {
	all, err := m.ListSecrets()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, n := range all {
		if strings.HasPrefix(n, prefix) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (m *Mock) ListKeys() ([]string, error) {
	if !m.initialized {
		return nil, spec.ErrInvalidVault
	}
	out := make([]string, 0, len(m.fingerprints))
	for _, fpr := range m.fingerprints {
		if _, ok := m.pubKeys[fpr]; ok {
			out = append(out, fpr)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (m *Mock) AddKey(armoredPublicKey string) error {
	if !m.initialized {
		return spec.ErrInvalidVault
	}
	if armoredPublicKey == "" {
		return fmt.Errorf("%w: empty armored key", spec.ErrInvalidKey)
	}
	fpr := Fingerprint(armoredPublicKey)
	if _, exists := m.pubKeys[fpr]; exists {
		return fmt.Errorf("%w: %s", spec.ErrDuplicateRecipient, fpr)
	}
	m.pubKeys[fpr] = armoredPublicKey
	m.fingerprints = append(m.fingerprints, fpr)
	sort.Strings(m.fingerprints)
	// Re-encrypt: in mock, update the recipient list on every existing
	// secret. Real backends must rewrite the .gpg ciphertext.
	for name := range m.secretMeta {
		m.secretMeta[name] = append([]string(nil), m.fingerprints...)
	}
	return nil
}

func (m *Mock) RemoveKey(fingerprint string) error {
	if !m.initialized {
		return spec.ErrInvalidVault
	}
	if len(m.fingerprints) <= 1 {
		return spec.ErrLastRecipient
	}
	idx := -1
	for i, f := range m.fingerprints {
		if f == fingerprint {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("%w: %s", spec.ErrRecipientNotFound, fingerprint)
	}
	m.fingerprints = append(m.fingerprints[:idx], m.fingerprints[idx+1:]...)
	delete(m.pubKeys, fingerprint)
	delete(m.privKeys, fingerprint)
	// Re-encrypt: in mock, update the recipient list on every existing
	// secret. Real backends must rewrite the .gpg ciphertext.
	for name := range m.secretMeta {
		m.secretMeta[name] = append([]string(nil), m.fingerprints...)
	}
	return nil
}
