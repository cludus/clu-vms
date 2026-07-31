package core

import (
	"clu-vms/internal/spec"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type QemuServiceImpl struct {
	dir string
}

func (qs *QemuServiceImpl) CreateVM(name string) (spec.QemuVM, error) {
	_, err := os.Stat(qs.dir + "/" + name + ".qcow2")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("vm file already exists: %w", err)
		}
	}

	source, err := os.Open(qs.dir + "/imgs/alpine.img")
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(qs.dir + "/" + name + ".qcow2")
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destination, source)
	if err != nil {
		return nil, fmt.Errorf("failed to copy data: %w", err)
	}

	// Commit the file contents to stable storage
	err = destination.Sync()
	if err != nil {
		return nil, fmt.Errorf("failed to sync destination file: %w", err)
	}

	return &QemuVMImpl{
		name: name,
		path: qs.dir + "/" + name + ".qcow2",
	}, nil
}

func (qs *QemuServiceImpl) OpenVM(name string) (spec.QemuVM, error) {
	return nil, nil
}

type QemuVMImpl struct {
	name string
	path string
}

func (qs *QemuVMImpl) Name() string {
	return qs.name
}

func (qs *QemuVMImpl) SetCloudInitConfig(config string) {

}

func (qs *QemuVMImpl) AddDataDisk(path string) {

}

func (qs *QemuVMImpl) AddNetworkDevice(bridge string, vlan int) {

}

func (qs *QemuVMImpl) Start() error {
	cmdArr := []string{"qemu-system-x86_64"}
	cmdArr = append(cmdArr, "-enable-kvm")
	cmdArr = append(cmdArr, "-m", "1G")
	cmdArr = append(cmdArr, "-cpu", "host")
	//cmdArr = append(cmdArr, "-daemonize", "-pidfile "+qs.name+".pid")
	cmdArr = append(cmdArr, "-drive", "file="+qs.path+",if=virtio,format=qcow2")
	cmdArr = append(cmdArr, "-netdev", "bridge,id=net0,br=vmbr0")
	cmdArr = append(cmdArr, "-device", "virtio-net-pci,netdev=net0")
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(output))
	return nil
}

func (qs *QemuVMImpl) Shutdown() error {
	return nil
}

func (qs *QemuVMImpl) IsRunning() bool {
	return false
}
