package conformance

import (
	"testing"

	"clu-vms/internal/impl/secrets/mock"
	"clu-vms/internal/spec"
)

// TestMockConformance verifies that the in-memory mock implementation
// satisfies the spec.VaultSecrets contract. Both real backends will run
// the same conformance.Run against their own factories.
func TestMockConformance(t *testing.T) {
	Run(t, func(t *testing.T) Backend {
		t.Helper()
		v := mock.New()
		dir := t.TempDir()
		if err := v.Init(dir); err != nil {
			t.Fatalf("mock Init: %v", err)
		}
		return Backend{
			Vault:       v,
			Dir:         dir,
			Fingerprint: mock.Fingerprint,
		}
	})
}

// Compile-time guarantee: mock satisfies spec.VaultSecrets (the suite
// above is the dynamic check; this catches API drift at build time).
var _ spec.VaultSecrets = (*mock.Mock)(nil)
