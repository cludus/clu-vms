package core

import (
	"clu-vms/internal/spec"
)

type FileHostHandler struct {
	workCtx  spec.WorkContext
	hostFile string
	host     spec.HostDefinition
}

func defaultHostDefinition() spec.HostDefinition {
	return spec.HostDefinition{
		Defaults: spec.VmDefaultDefinition{
			UserCredentials: spec.UserCredentials{
				User: "my-user",
				Pass: "my-pass",
			},
			OS:            spec.Alpine_3_24(),
			MemoryGB:      2,
			StorageSizeGB: 50,
		},
		Bridge: spec.HostBridge{
			Name:   "my-bridge",
			PhysIf: "vmbr0",
			IP:     "192.168.100.1",
			Create: false,
		},
		Hosts: []spec.VmDefinition{
			{
				Name:      "vm1",
				IPAddress: "192.168.100.10",
			},
		},
	}
}

func NewFileHostHandler(workCtx spec.WorkContext, fileName string) *FileHostHandler {
	hostFile := fileName
	if hostFile == "" {
		hostFile = "vms.yml"
	}

	host := defaultHostDefinition()
	if workCtx.FileExists(hostFile) {
		// read file to host definition
	}

	return &FileHostHandler{
		workCtx:  workCtx,
		hostFile: fileName,
		host:     host,
	}
}
