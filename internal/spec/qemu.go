package spec

type QemuService interface {
	CreateVM(name string) (QemuVM, error)
	OpenVM(name string) (QemuVM, error)
}

type QemuVM interface {
	Name() string
	SetCloudInitConfig(config string)
	AddDataDisk(path string, sizeGB int)
	AddNetworkDevice(bridge string, vlan int)
	Start() error
	Shutdown() error
	IsRunning() bool
}
