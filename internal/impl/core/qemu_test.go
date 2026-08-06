package core

import "testing"

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
		return
	}
}
