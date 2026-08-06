package core

import (
	"clu-vms/internal/spec"
	"fmt"
)

type QemuServiceImpl struct {
	workCtx spec.WorkContext
	imgs    spec.OSImages
	disks   spec.DiskService
	net     spec.NetworkService
}

func NewQemuServiceImpl(workCtx spec.WorkContext) spec.QemuService {
	return &QemuServiceImpl{
		workCtx: workCtx,
		imgs:    NewOSImages(workCtx),
		disks:   NewDiskService(workCtx),
		net:     NewNetworkService(),
	}
}

func (qs *QemuServiceImpl) CreateVM(name string) (spec.QemuVM, error) {
	vmFileName := fmt.Sprintf(".assets/%s/system.qcow2", name)
	if qs.workCtx.FileExists(vmFileName) {
		err := qs.workCtx.DeleteFile(vmFileName)
		if err != nil {
			return nil, err
		}
	}

	err := qs.imgs.CopyOSImage(spec.Alpine_3_24, vmFileName)
	if err != nil {
		return nil, err
	}

	return newQemuVM(name, qs.workCtx, qs.disks), nil
}

func (qs *QemuServiceImpl) OpenVM(name string) (spec.QemuVM, error) {
	return newQemuVM(name, qs.workCtx, qs.disks), nil
}

type qemuVMImpl struct {
	name     string
	workCtx  spec.WorkContext
	memoryGb int
	diskServ spec.DiskService
	disks    []string
	netDev   []string
}

func newQemuVM(name string, workCtx spec.WorkContext, diskServ spec.DiskService) *qemuVMImpl {
	return &qemuVMImpl{
		name:     name,
		workCtx:  workCtx,
		diskServ: diskServ,
		memoryGb: 1,
		disks:    []string{},
		netDev:   []string{},
	}
}

func (qs *qemuVMImpl) Name() string {
	return qs.name
}

func (qs *qemuVMImpl) SetCloudInitConfig(config string) {

}

func (qs *qemuVMImpl) AddDataDisk(path string, sizeGB int) {
	if !qs.diskServ.DataDiskExists(path) {
		err := qs.diskServ.CreateDataDisk(path, sizeGB, spec.FileSystemExt4)
		if err != nil {
			return
		}
	}
	qs.disks = append(qs.disks, path)
}

func (qs *qemuVMImpl) AddNetworkDevice(bridge string, vlan int) {
	if vlan != 0 {
		qs.netDev = append(qs.netDev, fmt.Sprintf("%s.%d", bridge, vlan))
	} else {
		qs.netDev = append(qs.netDev, bridge)
	}
}

func (qs *qemuVMImpl) Start() error {
	vmPath := fmt.Sprintf("%s/.assets/%s/system.qcow2", qs.workCtx.WorkDir(), qs.name)
	cmdArr := []string{"qemu-system-x86_64"}
	cmdArr = append(cmdArr, "-enable-kvm")
	cmdArr = append(cmdArr, "-m", fmt.Sprintf("%dG", qs.memoryGb))
	cmdArr = append(cmdArr, "-cpu", "host")
	//cmdArr = append(cmdArr, "-daemonize", "-pidfile "+qs.name+".pid")
	cmdArr = append(cmdArr, "-drive", "file="+vmPath+",if=virtio,format=qcow2")
	for _, disk := range qs.disks {
		diskPath := fmt.Sprintf("%s/%s", qs.workCtx.WorkDir(), disk)
		cmdArr = append(cmdArr, "-drive", "file="+diskPath+",if=virtio,format=qcow2")
	}
	for i, netDev := range qs.netDev {
		id := fmt.Sprintf("net%d", i+1)
		cmdArr = append(cmdArr, "-netdev", "bridge,id="+id+",br="+netDev)
		cmdArr = append(cmdArr, "-device", "virtio-net-pci,netdev="+id)
	}
	err := qs.workCtx.RunCommand(cmdArr...)
	if err != nil {
		return err
	}
	return nil
}

func (qs *qemuVMImpl) Shutdown() error {
	return nil
}

func (qs *qemuVMImpl) IsRunning() bool {
	return false
}
