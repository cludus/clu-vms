Proposed interfaces

```go

type S3Service interface {
	DownloadVM(url string, path string)
	UploadVM(path string, url string)
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