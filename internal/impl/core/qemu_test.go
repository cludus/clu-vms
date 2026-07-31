package core

import "testing"

func TestQemu(t *testing.T) {
	serv := &QemuServiceImpl{
		dir: "testdata",
	}

	vm, err := serv.CreateVM("testvm")
	if err != nil {
		t.Fatal(err)
	}
	err = vm.Start()
	if err != nil {
		return
	}
}
