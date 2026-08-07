package spec

type UserCredentials struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type HostBridge struct {
	Name   string `yaml:"name"`
	PhysIf string `yaml:"physIf"`
	IP     string `yaml:"ip"`
	Create bool   `yaml:"create"`
}

type VmDefinition struct {
	UserCredentials UserCredentials `yaml:"userCredentials"`
	OS              OSImagesType    `yaml:"os"`
	MemoryGB        uint16          `yaml:"memoryGB"`
	StorageSizeGB   uint16          `yaml:"storageSizeGB"`
	Name            string          `yaml:"name"`
	IPAddress       string          `yaml:"ipAddress"`
}

type VmDefaultDefinition struct {
	UserCredentials UserCredentials `yaml:"userCredentials"`
	OS              OSImagesType    `yaml:"os"`
	MemoryGB        uint16          `yaml:"memoryGB"`
	StorageSizeGB   uint16          `yaml:"storageSizeGB"`
}

type HostDefinition struct {
	Defaults VmDefaultDefinition `yaml:"defaults"`
	Bridge   HostBridge          `yaml:"bridge"`
	Hosts    []VmDefinition      `yaml:"hosts"`
}

type HostHandler interface {
	Definition() HostDefinition
	SetDefaults(definition VmDefaultDefinition) error
	SetBridge(bridge HostBridge) error
	AddHost(definition VmDefinition, overwriteIfExists bool) error
	CheckDefinition() error
}
