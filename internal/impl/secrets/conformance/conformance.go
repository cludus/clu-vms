package conformance

// Package conformance provides a shared test suite that every
// spec.VaultSecrets implementation must pass.
//
// Backends wire themselves into the suite via the Factory type, which
// returns a fresh Backend (vault + temp dir + fingerprint function) for
// each subtest. The suite never inspects on-disk layout — it is purely
// behavioural — so every backend runs the same battery against its own
// in-memory state.

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"clu-vms/internal/spec"
)

// Backend is what a Factory produces for one subtest.
type Backend struct {
	// Vault is a fresh, initialized spec.VaultSecrets.
	Vault spec.VaultSecrets

	// Dir is the vault's root path on disk. Tests do not currently
	// inspect it, but it is exposed for backends that want to write
	// additional tests outside the conformance suite.
	Dir string

	// Fingerprint returns the canonical 40-char uppercase hex fingerprint
	// for an armored public key block. Each backend supplies its own
	// implementation (e.g. SHA-256 for mock, real OpenPGP for ProtonMail,
	// parsed key id for GPGME) so the suite stays fingerprint-agnostic.
	Fingerprint func(armoredPublicKey string) string
}

// Factory creates a fresh Backend for one subtest.
type Factory func(t *testing.T) Backend

// Run executes the conformance suite against any backend produced by
// factory. Each subtest gets its own Backend.
func Run(t *testing.T, factory Factory) {
	t.Helper()

	t.Run("Init", func(t *testing.T) {
		t.Run("SucceedsOnEmpty", func(t *testing.T) {
			b := factory(t)
			if b.Vault == nil {
				t.Fatal("factory returned nil vault")
			}
			keys, err := b.Vault.ListKeys()
			if err != nil {
				t.Fatalf("ListKeys: %v", err)
			}
			if len(keys) != 0 {
				t.Fatalf("ListKeys on fresh vault: got %v, want empty", keys)
			}
		})
	})

	t.Run("SetGet", func(t *testing.T) {
		t.Run("RoundTrip", func(t *testing.T) { testSetGetRoundTrip(t, factory) })
		t.Run("RequiresRecipient", func(t *testing.T) { testSetSecretRequiresRecipient(t, factory) })
		t.Run("UnknownName", func(t *testing.T) { testGetSecretUnknownName(t, factory) })
	})

	t.Run("List", func(t *testing.T) {
		t.Run("SecretsSorted", func(t *testing.T) { testListSecretsSorted(t, factory) })
		t.Run("SecretsWithPrefix", func(t *testing.T) { testListSecretsWithPrefix(t, factory) })
		t.Run("KeysSorted", func(t *testing.T) { testListKeysSorted(t, factory) })
	})

	t.Run("AddKey", func(t *testing.T) {
		t.Run("Duplicate", func(t *testing.T) { testAddKeyDuplicate(t, factory) })
		t.Run("ReEncrypts", func(t *testing.T) { testAddKeyReEncrypts(t, factory) })
	})

	t.Run("RemoveKey", func(t *testing.T) {
		t.Run("LastFails", func(t *testing.T) { testRemoveKeyLastFails(t, factory) })
		t.Run("UnknownFails", func(t *testing.T) { testRemoveKeyUnknownFails(t, factory) })
		t.Run("ReEncrypts", func(t *testing.T) { testRemoveKeyReEncrypts(t, factory) })
	})

	t.Run("Unlock", func(t *testing.T) {
		t.Run("AllowsGet", func(t *testing.T) { testUnlockAllowsGet(t, factory) })
	})

	t.Run("Passphrase", func(t *testing.T) {
		t.Run("LazyUnlock", func(t *testing.T) { testPassphraseLazyUnlock(t, factory) })
	})
}

// ---------- helpers ----------

type testVault struct {
	v           spec.VaultSecrets
	pub1, priv1 string
	fpr1        string
	pub2, priv2 string
	fpr2        string
	pub3, priv3 string
	fpr3        string
}

func newTestVault(t *testing.T, factory Factory) testVault {
	t.Helper()
	b := factory(t)
	pub1, priv1 := testKeyPair("alice")
	pub2, priv2 := testKeyPair("bob")
	pub3, priv3 := testKeyPair("carol")
	if err := b.Vault.AddKey(pub1); err != nil {
		t.Fatalf("AddKey alice: %v", err)
	}
	return testVault{
		v:    b.Vault,
		pub1: pub1, priv1: priv1, fpr1: b.Fingerprint(pub1),
		pub2: pub2, priv2: priv2, fpr2: b.Fingerprint(pub2),
		pub3: pub3, priv3: priv3, fpr3: b.Fingerprint(pub3),
	}
}

// testKeyPair returns (pub, priv) armored blocks that share the same
// fingerprint under any reasonable derivation. The conformance suite
// never inspects armor headers; it goes through Backend.Fingerprint to
// derive the fingerprint for each backend.
func testKeyPair(label string) (pub, priv string) {
	body := fmt.Sprintf("mock-key-%s", label)
	pub = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n" + body + "\n-----END PGP PUBLIC KEY BLOCK-----"
	priv = pub // identical content so SHA-256 (mock) and OpenPGP (real)
	// backends agree on the fingerprint.
	return
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------- individual tests ----------

func testSetGetRoundTrip(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.SetSecret("greeting", "hello"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	if err := tv.v.UnlockPrivateKey(tv.priv1, []byte("ignored")); err != nil {
		t.Fatalf("UnlockPrivateKey: %v", err)
	}
	got, err := tv.v.GetSecret("greeting")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got != "hello" {
		t.Fatalf("GetSecret: got %q, want %q", got, "hello")
	}
}

func testSetSecretRequiresRecipient(t *testing.T, factory Factory) {
	b := factory(t)
	err := b.Vault.SetSecret("x", "y")
	if !errors.Is(err, spec.ErrNoRecipients) {
		t.Fatalf("SetSecret with no recipients: got %v, want %v", err, spec.ErrNoRecipients)
	}
}

func testGetSecretUnknownName(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	_, err := tv.v.GetSecret("missing")
	if !errors.Is(err, spec.ErrSecretNotFound) {
		t.Fatalf("GetSecret unknown: got %v, want %v", err, spec.ErrSecretNotFound)
	}
}

func testListSecretsSorted(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	for _, n := range []string{"zeta", "alpha", "mu"} {
		if err := tv.v.SetSecret(n, n); err != nil {
			t.Fatalf("SetSecret %s: %v", n, err)
		}
	}
	got, err := tv.v.ListSecrets()
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	want := []string{"alpha", "mu", "zeta"}
	if !equalStrings(got, want) {
		t.Fatalf("ListSecrets: got %v, want %v", got, want)
	}
}

func testListSecretsWithPrefix(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	for _, n := range []string{"db/password", "db/user", "api/key", "misc"} {
		if err := tv.v.SetSecret(n, n); err != nil {
			t.Fatalf("SetSecret %s: %v", n, err)
		}
	}
	got, err := tv.v.ListSecretsWithPrefix("db/")
	if err != nil {
		t.Fatalf("ListSecretsWithPrefix: %v", err)
	}
	want := []string{"db/password", "db/user"}
	if !equalStrings(got, want) {
		t.Fatalf("ListSecretsWithPrefix: got %v, want %v", got, want)
	}
}

func testListKeysSorted(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.AddKey(tv.pub2); err != nil {
		t.Fatalf("AddKey bob: %v", err)
	}
	got, err := tv.v.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListKeys: got %v, want 2 entries", got)
	}
	if !sort.StringsAreSorted(got) {
		t.Fatalf("ListKeys not sorted: %v", got)
	}
}

func testAddKeyDuplicate(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	err := tv.v.AddKey(tv.pub1)
	if !errors.Is(err, spec.ErrDuplicateRecipient) {
		t.Fatalf("AddKey duplicate: got %v, want %v", err, spec.ErrDuplicateRecipient)
	}
}

func testAddKeyReEncrypts(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.SetSecret("s", "v1"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	if err := tv.v.AddKey(tv.pub2); err != nil {
		t.Fatalf("AddKey bob: %v", err)
	}
	// After AddKey, the secret is encrypted to both alice and bob. Bob
	// must be able to read it without alice being unlocked.
	if err := tv.v.UnlockPrivateKey(tv.priv2, []byte("ignored")); err != nil {
		t.Fatalf("UnlockPrivateKey bob: %v", err)
	}
	got, err := tv.v.GetSecret("s")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got != "v1" {
		t.Fatalf("GetSecret: got %q, want %q", got, "v1")
	}
}

func testRemoveKeyLastFails(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	err := tv.v.RemoveKey(tv.fpr1)
	if !errors.Is(err, spec.ErrLastRecipient) {
		t.Fatalf("RemoveKey last: got %v, want %v", err, spec.ErrLastRecipient)
	}
}

func testRemoveKeyUnknownFails(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.AddKey(tv.pub2); err != nil {
		t.Fatalf("AddKey bob: %v", err)
	}
	unknown := strings.Repeat("0", 40)
	err := tv.v.RemoveKey(unknown)
	if !errors.Is(err, spec.ErrRecipientNotFound) {
		t.Fatalf("RemoveKey unknown: got %v, want %v", err, spec.ErrRecipientNotFound)
	}
}

func testRemoveKeyReEncrypts(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.AddKey(tv.pub2); err != nil {
		t.Fatalf("AddKey bob: %v", err)
	}
	if err := tv.v.SetSecret("s", "v1"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	// Unlock bob only; alice stays locked.
	if err := tv.v.UnlockPrivateKey(tv.priv2, []byte("ignored")); err != nil {
		t.Fatalf("UnlockPrivateKey bob: %v", err)
	}
	if err := tv.v.RemoveKey(tv.fpr1); err != nil {
		t.Fatalf("RemoveKey alice: %v", err)
	}
	// Bob must still decrypt; alice is no longer in the recipient set.
	got, err := tv.v.GetSecret("s")
	if err != nil {
		t.Fatalf("GetSecret after RemoveKey: %v", err)
	}
	if got != "v1" {
		t.Fatalf("GetSecret: got %q, want %q", got, "v1")
	}
}

func testUnlockAllowsGet(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.SetSecret("s", "v"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	// Before unlock, GetSecret must fail with ErrDecryptionFailed.
	if _, err := tv.v.GetSecret("s"); !errors.Is(err, spec.ErrDecryptionFailed) {
		t.Fatalf("GetSecret before unlock: got %v, want %v", err, spec.ErrDecryptionFailed)
	}
	if err := tv.v.UnlockPrivateKey(tv.priv1, []byte("ignored")); err != nil {
		t.Fatalf("UnlockPrivateKey: %v", err)
	}
	got, err := tv.v.GetSecret("s")
	if err != nil {
		t.Fatalf("GetSecret after unlock: %v", err)
	}
	if got != "v" {
		t.Fatalf("GetSecret: got %q, want %q", got, "v")
	}
}

func testPassphraseLazyUnlock(t *testing.T, factory Factory) {
	tv := newTestVault(t, factory)
	if err := tv.v.SetSecret("s", "v"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	called := 0
	tv.v.SetPassphraseFunc(func(fpr string) ([]byte, error) {
		called++
		return []byte("ignored"), nil
	})
	got, err := tv.v.GetSecret("s")
	if err != nil {
		t.Fatalf("GetSecret with callback: %v", err)
	}
	if got != "v" {
		t.Fatalf("GetSecret: got %q, want %q", got, "v")
	}
	if called == 0 {
		t.Fatal("passphrase callback was not invoked")
	}
	// Second call should not re-invoke the callback (mock caches; real
	// backends should also avoid re-prompting within the same session).
	prev := called
	if _, err := tv.v.GetSecret("s"); err != nil {
		t.Fatalf("GetSecret (cached): %v", err)
	}
	if called != prev {
		t.Fatalf("callback invoked %d times on second GetSecret, want %d (cache miss)", called-prev, 0)
	}
}
