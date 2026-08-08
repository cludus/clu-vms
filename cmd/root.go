package cmd

import (
	"clu-vms/internal/cli/host"
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

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "vms.yml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddGroup(&cobra.Group{ID: "host", Title: "Host Commands"})

	host.RootCmd.GroupID = "host"
	rootCmd.AddCommand(host.RootCmd)
}
