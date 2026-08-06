package spec

type NetworkService interface {
	CreateBridge(bridge string, ipAddr string, iface string) error
	DeleteBridge(bridge string, ipAddr string, iface string) error
	CreateVlan(bridge string, vlan int) error
	DeleteVlan(bridge string, vlan int) error
}
