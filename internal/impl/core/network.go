package core

import "clu-vms/internal/spec"

type networkServiceImpl struct {
}

func NewNetworkService() spec.NetworkService {
	return &networkServiceImpl{}
}

func (ns *networkServiceImpl) CreateBridge(bridge string, ipAddr string, iface string) error {
	return nil
}

func (ns *networkServiceImpl) DeleteBridge(bridge string, ipAddr string, iface string) error {
	return nil
}

func (ns *networkServiceImpl) CreateVlan(bridge string, vlan int) error {
	return nil
}

func (ns *networkServiceImpl) DeleteVlan(bridge string, vlan int) error {
	return nil
}
