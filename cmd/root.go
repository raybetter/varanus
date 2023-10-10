package cmd

import (
	"varanus/internal/app"

	"github.com/spf13/cobra"
)

// CmdContext provides a structure for all the dependencies we pass into the command structures
type CmdContext struct {
	App app.VaranusApp
}

// rootCmd represents the base command when called without any subcommands
func MakeRootCmd(context *CmdContext) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "varanus",
		Short: "An automated server monitoring tool.",
		Long: `Define a configuration YAML file, seal it, then run the server to begin monitoring
email servers.
`,
	}

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.varanus.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//register subcommands
	configCmd := makeConfigCmd(context)
	rootCmd.AddCommand(configCmd)

	return rootCmd
}
