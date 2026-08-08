package core

import (
	"testing"
	"time"
)

func TestQemu(t *testing.T) {
	workCtx, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}
	serv := NewQemuServiceImpl(workCtx)

	vm, err := serv.CreateVM("testvm")
	if err != nil {
		t.Fatal(err)
	}

	vm.AddNetworkDevice("vmbr0", 0)
	vm.AddDataDisk("disk.img", 10)

	err = vm.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)
	err = vm.Shutdown()
	if err != nil {
		t.Fatal(err)
	}
	for vm.IsRunning() {
		time.Sleep(1 * time.Second)
	}
}
