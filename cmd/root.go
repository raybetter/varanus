/*
Copyright © 2023 Justin Ray

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
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
		Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
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
