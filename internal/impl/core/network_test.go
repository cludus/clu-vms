package core

import "testing"

func TestNetworkService(t *testing.T) {
	netServ := NewNetworkService()
	err := netServ.CreateBridge("br0", "192.168.0.100", "eth0")
	if err != nil {
		return
	}
}
