package spec

type QemuService interface {
	CreateVM(name string) QemuVM
	OpenVM(name string) QemuVM
}

type QemuVM interface {
	Name() string
	SetCloudInitConfig(config string)
	AddDataDisk(path string)
	AddNetworkDevice(bridge string, vlan int)
	Start() error
	Shutdown() error
	IsRunning() bool
}
