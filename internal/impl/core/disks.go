package core

import (
	"clu-vms/internal/spec"
	"fmt"
)

type DiskServiceImpl struct {
	workCtx spec.WorkContext
}

func NewDiskService(wc spec.WorkContext) spec.DiskService {
	return &DiskServiceImpl{
		workCtx: wc,
	}
}

func (d *DiskServiceImpl) DataDiskExists(path string) bool {
	return d.workCtx.FileExists(path)
}

func (d *DiskServiceImpl) CreateDataDisk(path string, sizeGB int, fs spec.FileSystem) error {
	//qemu-img create -f qcow2 data/system.img 100G
	err := d.workCtx.RunCommand("qemu-img", "create", "-f", "qcow2", path, fmt.Sprintf("%dG", sizeGB))
	if err != nil {
		return err
	}
	return nil
}

func (d *DiskServiceImpl) DeleteDataDisk(path string) error {
	return d.workCtx.DeleteFile(path)
}
