package core

import (
	"clu-vms/internal/spec"
	"testing"
)

func TestDiskService_CreateDataDisk(t *testing.T) {
	workCtx, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}
	diskService := NewDiskService(workCtx)
	err = diskService.CreateDataDisk("my-data.qcow2", 10, spec.FileSystemExt4)
	if err != nil {
		t.Fatal(err)
	}
}
