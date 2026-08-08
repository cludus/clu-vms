package host

import (
	"clu-vms/internal/impl/core"
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

var checkHostDefinitionCmd = &cobra.Command{
	Use:     "check-definition",
	Aliases: []string{"check-def"},
	Short:   "Check the host definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		ctx, err := core.NewLocalContext(wd)
		if err != nil {
			return err
		}

		cfgFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		handler, err := core.NewFileHostHandler(ctx, &cfgFile)
		if err != nil {
			return err
		}

		cmd.Printf("Checking host definition for file %s...\n", cfgFile)
		err = handler.CheckDefinition()
		if err != nil {
			cmd.Printf("  Check failed: %v\n", err)
		} else {
			cmd.Println("  Host definition is valid!")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(checkHostDefinitionCmd)
}
