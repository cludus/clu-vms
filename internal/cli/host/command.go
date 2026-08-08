package host

import (
	"clu-vms/internal/impl/core"
	"clu-vms/internal/spec"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "host",
	Aliases: []string{"h"},
	Short:   "Manage your virtual machines on the host",
	Long: `The host command allows you to manage your virtual machines on the host system.
It provides a set of subcommands to check host definitions and manage virtual machines.`,
}

func createHostHandler(cmd *cobra.Command) (spec.HostHandler, spec.WorkContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	ctx, err := core.NewLocalContext(wd)
	if err != nil {
		return nil, nil, err
	}

	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, nil, err
	}

	handler, err := core.NewFileHostHandler(ctx, &cfgFile)
	if err != nil {
		return nil, nil, err
	}
	return handler, ctx, nil
}

var checkHostDefinitionCmd = &cobra.Command{
	Use:   "check",
	Args:  cobra.NoArgs,
	Short: "Check the host definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, _, err := createHostHandler(cmd)
		if err != nil {
			return err
		}

		cmd.Println("Checking host definition...")
		err = handler.CheckDefinition()
		if err != nil {
			cmd.Printf("  Check failed: %v\n", err)
		} else {
			cmd.Println("  Host definition is valid!")
		}

		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start host vms",
	Long:  "Start all host vms in the host system, or a single vm by name",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, ctx, err := createHostHandler(cmd)
		if err != nil {
			return err
		}

		cmd.Println("Checking host definition...")
		err = handler.CheckDefinition()
		if err != nil {
			cmd.Printf("  Check failed: %v\n", err)
		}

		serv := core.NewQemuServiceImpl(ctx)
		// TODO: check bridge exists

		if len(args) > 0 {
			return startVm(args[0], handler, serv, cmd)
		}
		return startAllVm(handler, serv, cmd)
	},
}

func init() {
	RootCmd.AddCommand(checkHostDefinitionCmd)
	RootCmd.AddCommand(startCmd)
}

func startVm(name string, handler spec.HostHandler, serv spec.QemuService, cmd *cobra.Command) error {
	host, err := handler.Host(name)
	if err != nil {
		return err
	}

	return startHost(host, handler, serv, cmd)
}

func startAllVm(handler spec.HostHandler, serv spec.QemuService, cmd *cobra.Command) error {
	for _, host := range handler.Definition().Hosts {
		err := startHost(&host, handler, serv, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func startHost(definition *spec.VmDefinition, handler spec.HostHandler, serv spec.QemuService, cmd *cobra.Command) error {
	cmd.Printf("Starting host %s...\n", definition.Name)
	vm, err := serv.CreateVM(definition.Name)
	if err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	hostDefinition := handler.Definition()
	if verbose {
		cmd.Printf("  Using bridge %s\n", hostDefinition.Bridge.Name)
	}
	vm.AddNetworkDevice(hostDefinition.Bridge.Name, 0)

	// TODO: create disk if not exists
	// vm.AddDataDisk(disk.Path, disk.SizeGB)

	err = vm.Start()
	if err != nil {
		return err
	}

	cmd.Println("  Host started successfully!")
	return nil
}
