package core

import (
	"clu-vms/internal/spec"
	"testing"
)

func TestWorkContext(t *testing.T) {
	wc := NewLocalContext("/tmp")
	if wc.GetType() != spec.WorkContextTypeLocal {
		t.Fatal("wrong work context type")
	}

	wc.CreateDir("/tmp/test")
	//TODO test all methods
	t.Fatal("TODO not implemented")
}
