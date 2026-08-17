package spec

// VaultSecrets defines the contract for a GPG-backed secret vault.
//
// Storage layout (set up by Init):
//
//	<root>/
//	  .keys.json          # {"recipients": ["<fpr1>", "<fpr2>"]}
//	  keys/<fpr>.asc      # ASCII-armored public key per recipient
//	  secrets/<name>.gpg  # one armored OpenPGP message per secret
//
// Invariants:
//   - At least one recipient is always configured. RemoveKey on the last
//     recipient is a hard error.
//   - Only public keys live on disk. Private keys are loaded into process
//     memory for the duration of a session and never persisted by the vault.
//
// A secret is a single OpenPGP message encrypted to the current recipient
// set. AddKey and RemoveKey therefore rewrite every existing secret so it
// is encrypted to the new roster.
type VaultSecrets interface {
	// Lifecycle

	// Init prepares the vault rooted at path. It creates the directory
	// tree, loads .keys.json, and rebuilds the in-memory fingerprint
	// cache from keys/<fpr>.asc. If the vault is empty, Init still
	// succeeds but every operation that needs a recipient will fail
	// until AddKey has been called.
	Init(path string) error

	// Decryption bootstrap. Private keys are never persisted by the
	// vault, so GetSecret cannot succeed unless at least one private
	// key has been unlocked or a passphrase callback is registered.

	// UnlockPrivateKey parses an armored private key block and unlocks
	// it in process memory using passphrase. The unlocked entity is
	// retained for the lifetime of the VaultSecrets instance.
	//
	// For the gpgme backend this is a no-op: private keys are kept by
	// gpg-agent and the passphrase is requested via pinentry.
	UnlockPrivateKey(armoredKey string, passphrase []byte) error

	// SetPassphraseFunc installs a callback used by GetSecret to obtain
	// the passphrase for a private key identified by fingerprint. This
	// enables lazy unlocking of keys that were not pre-loaded via
	// UnlockPrivateKey.
	//
	// For the gpgme backend the callback is optional and only useful
	// when overriding pinentry mode.
	SetPassphraseFunc(fn func(fingerprint string) ([]byte, error))

	// Secrets

	// GetSecret returns the plaintext value of the named secret.
	// Requires a recipient whose private key is available (either
	// unlocked or reachable via SetPassphraseFunc).
	GetSecret(name string) (string, error)

	// SetSecret encrypts value to all current recipients and writes it
	// to secrets/<name>.gpg. Refuses if the recipient set is empty.
	SetSecret(name, value string) error

	// ListSecrets returns the names of every secret in the vault,
	// sorted, without the .gpg suffix.
	ListSecrets() ([]string, error)

	// ListSecretsWithPrefix returns the names of secrets whose name
	// starts with prefix, sorted, without the .gpg suffix.
	ListSecretsWithPrefix(prefix string) ([]string, error)

	// Recipients

	// ListKeys returns the fingerprints of every configured recipient.
	ListKeys() ([]string, error)

	// AddKey imports a recipient from an ASCII-armored public key
	// block, persists keys/<fpr>.asc, updates .keys.json, and rewrites
	// every existing secret so it is encrypted to the new recipient
	// set. Refuses if a recipient with the same fingerprint is already
	// configured.
	AddKey(armoredPublicKey string) error

	// RemoveKey drops the recipient with the given fingerprint, deletes
	// keys/<fpr>.asc, updates .keys.json, and rewrites every existing
	// secret so it is encrypted to the remaining recipients. Refuses
	// if fingerprint is the last configured recipient.
	//
	// Note: this stops future encryption to that recipient but does not
	// retroactively revoke historical messages. Anyone who still holds
	// the removed recipient's private key can still decrypt messages
	// encrypted before the removal.
	RemoveKey(fingerprint string) error
}
