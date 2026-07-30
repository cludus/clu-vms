Proposed interfaces

```go

type S3Service interface {
	DownloadVM(url string, path string)
	UploadVM(path string, url string)
}

type FileSystem int

const (
	FileSystemExt4 FileSystem = iota
	FileSystemXFS
)

type ImagesService interface {
	DownloadVMImage(url string, path string) error
	ConvertToQcow2(path string) error
}

type DiskService interface {
	CreateDataDisk(path string, sizeGB int, fs FileSystem) error
}

type NetworkService interface {
	CreateBridge(bridge string, ipAddr string, iface string) error
	DeleteBridge(bridge string, ipAddr string, iface string) error
	CreateVlan(bridge string, vlan int) error
	DeleteVlan(bridge string, vlan int) error
}

type CloudInitService interface {
	CreateCloudInitConfig(name string) CloudInitConfig
}

type CloudInitConfig interface {
	Name() string
	AddUser(name string, password string)
	AddSSHKey(key string)
	AddPackage(name string)
	AddNetworkConfig(iface string, network string, gateway string, dns string)
	AddFile(path string, content string)
	AddMountDisk(device string, mountPoint string)
	Build() string
}

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

type RouterVM interface {
	AddOSPFNetwork(network string, gateway string, mask string)
	AddBGPNetwork(network string, gateway string, mask string)
}

type DockerVM interface {
	RunCompose(context string, content string)
	AddConfigFile(context string, path string, content string)
}

type KubernetesVM interface {
}

```