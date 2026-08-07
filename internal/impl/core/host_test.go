package core

import (
	"clu-vms/internal/spec"
	"testing"
)

func TestHostParsing(t *testing.T) {
	workCtx, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}

	hostHandler, err := NewFileHostHandler(workCtx, nil)
	if err != nil {
		t.Fatal(err)
	}

	if hostHandler.host.Defaults.UserCredentials.User != "alpine" {
		t.Errorf("expected user to be 'alpine', got %q", hostHandler.host.Defaults.UserCredentials.User)
	}
	if hostHandler.host.Defaults.UserCredentials.Pass != "mypass1" {
		t.Errorf("expected pass to be 'mypass1', got %q", hostHandler.host.Defaults.UserCredentials.Pass)
	}
	if hostHandler.host.Defaults.OS != spec.Alpine_3_24 {
		t.Errorf("expected OS to be 'alpine_3_24', got %q", hostHandler.host.Defaults.OS)
	}
	if hostHandler.host.Defaults.MemoryGB != 4 {
		t.Errorf("expected MemoryGB to be 4, got %d", hostHandler.host.Defaults.MemoryGB)
	}
	if hostHandler.host.Defaults.StorageSizeGB != 100 {
		t.Errorf("expected StorageSizeGB to be 100, got %d", hostHandler.host.Defaults.StorageSizeGB)
	}

	if len(hostHandler.host.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hostHandler.host.Hosts))
	}
	if hostHandler.host.Hosts[0].Name != "vm-first" {
		t.Errorf("expected first host name to be 'vm-first', got %q", hostHandler.host.Hosts[0].Name)
	}
	if hostHandler.host.Hosts[0].IPAddress != "192.168.100.50" {
		t.Errorf("expected first host IP address to be '192.168.100.50', got %q", hostHandler.host.Hosts[0].IPAddress)
	}
}

func TestFileHostHandlerDefinitionAndSetDefaults(t *testing.T) {
	workCtx, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}

	testFile := "host-test-temp.yml"
	defer workCtx.DeleteFile(testFile)

	hostHandler, err := NewFileHostHandler(workCtx, &testFile)
	if err != nil {
		t.Fatal(err)
	}

	definition := hostHandler.Definition()
	if definition.Defaults.UserCredentials.User != "my-user" {
		t.Errorf("expected default user 'my-user', got %q", definition.Defaults.UserCredentials.User)
	}
	if len(definition.Hosts) != 1 || definition.Hosts[0].Name != "vm1" {
		t.Errorf("expected default host 'vm1', got %+v", definition.Hosts)
	}

	newDefaults := spec.VmDefaultDefinition{
		UserCredentials: spec.UserCredentials{
			User: "test-user",
			Pass: "test-pass",
		},
		OS:            spec.Alpine_3_24,
		MemoryGB:      8,
		StorageSizeGB: 200,
	}
	err = hostHandler.SetDefaults(newDefaults)
	if err != nil {
		t.Fatalf("SetDefaults failed: %v", err)
	}

	if hostHandler.Definition().Defaults.UserCredentials.User != "test-user" {
		t.Errorf("expected updated user 'test-user', got %q", hostHandler.Definition().Defaults.UserCredentials.User)
	}
	if hostHandler.Definition().Defaults.MemoryGB != 8 {
		t.Errorf("expected updated MemoryGB 8, got %d", hostHandler.Definition().Defaults.MemoryGB)
	}

	reloadedHandler, err := NewFileHostHandler(workCtx, &testFile)
	if err != nil {
		t.Fatal(err)
	}
	if reloadedHandler.Definition().Defaults.UserCredentials.Pass != "test-pass" {
		t.Errorf("expected persisted pass 'test-pass', got %q", reloadedHandler.Definition().Defaults.UserCredentials.Pass)
	}
}
