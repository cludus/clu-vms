package spec

type UserCredentials struct {
	User string
	Pass string
}

type HostBridge struct {
	Name   string
	PhysIf string
	IP     string
	Create bool
}

type VmDefinition struct {
	UserCredentials UserCredentials
	OS              OSImagesType
	MemoryGB        uint16
	StorageSizeGB   uint16
	Name            string
	IPAddress       string
}

type VmDefaultDefinition struct {
	UserCredentials UserCredentials
	OS              OSImagesType
	MemoryGB        uint16
	StorageSizeGB   uint16
}

type HostDefinition struct {
	Defaults VmDefaultDefinition
	Bridge   HostBridge
	Hosts    []VmDefinition
}

type HostHandler interface {
	Definition() HostDefinition
	SetDefaults(definition VmDefaultDefinition) error
	SetBridge(bridge HostBridge) error
	AddHost(definition VmDefinition) error
}
