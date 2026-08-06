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
