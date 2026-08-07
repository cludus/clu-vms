package cmd

import (
	"clu-vms/internal/impl/core"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cvms",
	Short: "cvms CLI allows you to manage your virtual machines",
	Long: `cvms is a command line interface (CLI) tool that allows you to manage your virtual machines.
It provides a set of commands to create, start, stop, and delete virtual machines, as well as manage their configurations and resources.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile string
var verbose bool

var checkHostDefinitionCmd = &cobra.Command{
	Use:   "host-check-definition",
	Short: "Check the host definition",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "vms.yml", "config file (default is vms.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(checkHostDefinitionCmd)
}
