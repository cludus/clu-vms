package core

import (
	"clu-vms/internal/spec"
	"clu-vms/internal/utils"
	"fmt"
)

type FileHostHandler struct {
	workCtx     spec.WorkContext
	hostFile    string
	host        spec.HostDefinition
	isDraftHost bool
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

	err = handler.workCtx.WriteFile(handler.hostFile, content)
	if err != nil {
		return err
	}

	handler.isDraftHost = false
	return nil
}

func NewFileHostHandler(workCtx spec.WorkContext, fileName *string) (spec.HostHandler, error) {
	hostFile := ""
	if fileName != nil {
		hostFile = *fileName
	}
	if hostFile == "" {
		hostFile = "vms.yml"
	}

	var host spec.HostDefinition
	var isDraftHost bool
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
		isDraftHost = true
	}

	return &FileHostHandler{
		workCtx:     workCtx,
		hostFile:    hostFile,
		host:        host,
		isDraftHost: isDraftHost,
	}, nil
}

func (handler *FileHostHandler) CheckDefinition() error {
	// Check if it is a draft host definition
	if handler.isDraftHost {
		return fmt.Errorf("host definition file %q does not exist, modify your host definition and save your changes", handler.hostFile)
	}

	// Check bridge definition
	if handler.host.Bridge.Name == "" {
		return fmt.Errorf("bridge name is missing")
	}
	if handler.host.Bridge.PhysIf == "" {
		return fmt.Errorf("bridge physical interface is missing")
	}
	if handler.host.Bridge.IP == "" {
		return fmt.Errorf("bridge IP address is missing")
	}

	hasDefaultsCredentials := handler.host.Defaults.UserCredentials.User != "" && handler.host.Defaults.UserCredentials.Pass != ""
	hasDefaultsOS := handler.host.Defaults.OS != ""

	// Check each host definition
	for _, host := range handler.host.Hosts {
		if host.Name == "" {
			return fmt.Errorf("host name is missing")
		}
		if host.IPAddress == "" {
			return fmt.Errorf("host %q IP address is missing", host.Name)
		}
		if !hasDefaultsCredentials && (host.UserCredentials.User == "" || host.UserCredentials.Pass == "") {
			return fmt.Errorf("host %q user credentials are missing", host.Name)
		}
		if !hasDefaultsOS && host.OS == "" {
			return fmt.Errorf("host %q OS is missing", host.Name)
		}
	}

	return nil
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
		return fmt.Errorf("host %q already exists", definition.Name)
	}

	handler.host.Hosts = append(handler.host.Hosts, definition)
	return handler.save()
}

func (handler *FileHostHandler) DeleteHost(name string) error {
	hostCurrentIndex := -1
	for i, host := range handler.host.Hosts {
		if host.Name == name {
			hostCurrentIndex = i
			break
		}
	}
	if hostCurrentIndex != -1 {
		handler.host.Hosts = append(handler.host.Hosts[:hostCurrentIndex], handler.host.Hosts[hostCurrentIndex+1:]...)
		return handler.save()
	}

	return fmt.Errorf("host %q does not exist", name)
}

func (handler *FileHostHandler) Start(name ...string) error {
	return fmt.Errorf("Start is not implemented")
}

func (handler *FileHostHandler) Stop(name ...string) error {
	return fmt.Errorf("Stop is not implemented")
}
