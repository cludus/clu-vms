package core

import (
	"clu-vms/internal/spec"
	"clu-vms/internal/utils"
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
			OS:            spec.Alpine_3_24,
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

func (handler *FileHostHandler) save() error {
	content, err := utils.SerializeYaml(handler.host)
	if err != nil {
		return err
	}

	return handler.workCtx.WriteFile(handler.hostFile, content)
}

func NewFileHostHandler(workCtx spec.WorkContext, fileName *string) (*FileHostHandler, error) {
	hostFile := ""
	if fileName != nil {
		hostFile = *fileName
	}
	if hostFile == "" {
		hostFile = "vms.yml"
	}

	var host spec.HostDefinition
	if workCtx.FileExists(hostFile) {
		content, err := workCtx.ReadFile(hostFile)
		if err != nil {
			return nil, err
		}

		err = utils.ParseYaml(content, &host)
		if err != nil {
			return nil, err
		}
	} else {
		host = defaultHostDefinition()
	}

	return &FileHostHandler{
		workCtx:  workCtx,
		hostFile: hostFile,
		host:     host,
	}, nil
}

func (handler *FileHostHandler) Definition() spec.HostDefinition {
	return handler.host
}

func (handler *FileHostHandler) SetDefaults(definition spec.VmDefaultDefinition) error {
	handler.host.Defaults = definition
	return handler.save()
}

func (handler *FileHostHandler) SetBridge(bridge spec.HostBridge) error {
	handler.host.Bridge = bridge
	return handler.save()
}

func (handler *FileHostHandler) AddHost(definition spec.VmDefinition, overwriteIfExists bool) error {
	hostCurrentIndex := -1
	for i, host := range handler.host.Hosts {
		if host.Name == definition.Name {
			hostCurrentIndex = i
			break
		}
	}
	if overwriteIfExists && hostCurrentIndex != -1 {
		handler.host.Hosts[hostCurrentIndex] = definition
		return handler.save()
	}
	if hostCurrentIndex != -1 {
		return &spec.HostAlreadyExistsError{Name: definition.Name}
	}

	handler.host.Hosts = append(handler.host.Hosts, definition)
	return handler.save()
}
