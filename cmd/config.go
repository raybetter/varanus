/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"strings"

	"varanus/internal/app"

	"github.com/spf13/cobra"
)

func makeConfigCmd(context *CmdContext) *cobra.Command {

	// configCmd represents the config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Operations for YAML configs",
		Long:  `Operations for YAML configs`,
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//register subcommands
	sealCmd := makeSealCmd(context)
	configCmd.AddCommand(sealCmd)

	unsealCmd := makeUnsealCmd(context)
	configCmd.AddCommand(unsealCmd)

	checkCmd := makeCheckCmd(context)
	configCmd.AddCommand(checkCmd)

	return configCmd
}

func makeCheckCmd(context *CmdContext) *cobra.Command {

	var cmdArgs = app.CheckConfigArgs{}

	// cmd represents the check command
	var cmd = &cobra.Command{
		Use:   "check",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			context.App.CheckConfig(&cmdArgs, os.Stdout)
		},
	}

	// Here you will define your flags and configuration settings.

	//local flags
	cmdArgs.Input = cmd.Flags().StringP("input", "i", "", "The filename of the YAML config to be unsealed.")
	cmd.MarkFlagRequired("input")
	cmd.MarkFlagFilename("input", "yaml", "yml")

	cmdArgs.PrivateKey = cmd.Flags().StringP("privateKey", "k", "", "The filename of the private key used to seal the config.")
	cmd.MarkFlagFilename("privateKey")

	cmdArgs.Passphrase = cmd.Flags().StringP("passphrase", "p", "", "The passphrase for the private key, if there is one.")

	return cmd

}

const SEALED_FILE_TOKEN = "sealed"

func makeSealCmd(context *CmdContext) *cobra.Command {

	cmdArgs := app.SealConfigArgs{}

	cmd := &cobra.Command{
		Use:   "seal",
		Short: "Seal the sensitive values of a configuration with a public key",
		Long: `Seal parses a config file and replaces sensitive values, such as passwords, with RSA
	encrypted values.  Since asymmetric encryption is used, only the public key is required for sealing.
	
	The corresponding private key must be provided to the varanus application using the config so that
	it can unseal the config values during monitoring.
	
	For example, consider the following YAML file:
	
	  mail:
	  accounts:
		- name: test1
		smtp:
		  sender_address: "example@example.com"
		  server_address: "smtp.example.com"
		  port: 465
		  username: joeuser@example.com
		  password: it's a secret
		imap:
		  server_address: "imap.example.com"
		  port: 993
		  username: janeuser@example.com
		  password: it's a secret
	  send_limits: []
	
	After running the seal command, the output file will look like:
	
	mail:
	  accounts:
		- name: test1
		  smtp:
			sender_address: "example@example.com"
			server_address: "smtp.example.com"	
			port: 465
			username: joeuser@example.com
			password: sealed(<encrypted string>)
		  imap:
			server_address: "imap.example.com"
			port: 993
			username: janeuser@example.com
			password: sealed(<encrypted string>)
	  send_limits: []
	
	Repeated calls to seal() will ignore values that are already sealed and only seal any unsealed
	values. Ensure that you are using the same private key, or the resulting config file will not work
	because the varanus server only accepts a single key.
	`,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := resolveAndCheckSealArgs(&cmdArgs)
			if err != nil {
				return err
			}

			//once we get through validation, silence the usage
			cmd.SilenceUsage = true

			err = context.App.SealConfig(&cmdArgs, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// Here you will define your flags and configuration settings.

	//local flags
	cmdArgs.Input = cmd.Flags().StringP("input", "i", "", "The filename of the YAML config to be sealed.")
	cmd.MarkFlagRequired("input")
	cmd.MarkFlagFilename("input", "yaml", "yml")

	cmdArgs.PublicKey = cmd.Flags().StringP("publicKey", "k", "", "The filename of the public key used to seal the config.")
	cmd.MarkFlagRequired("publicKey")
	cmd.MarkFlagFilename("publicKey")

	cmdArgs.Output = cmd.Flags().StringP("output", "o", "", "The filename to write the output to.  If omitted, the input file path is used with '.sealed' injected before the extension.")
	cmd.MarkFlagFilename("output")

	cmdArgs.ForceOverwrite = cmd.Flags().BoolP("forceOverwrite", "f", false, "If set, overwrite an existing file with the output.")

	return cmd

}

func injectValueBeforeExtension(filename string, token string) string {
	//set output file from input file
	tokens := strings.Split(filename, ".")
	if len(tokens) == 1 {
		//no extensions found
		return filename + "." + token
	} else {
		//output file
		tokens = append(tokens, tokens[len(tokens)-1]) //add the last token to the end again
		tokens[len(tokens)-2] = token                  //set the old last token to the sealed token
		return strings.Join(tokens, ".")
	}

}

func resolveAndCheckSealArgs(args *app.SealConfigArgs) error {
	if *args.Output == "" {
		*args.Output = injectValueBeforeExtension(*args.Input, SEALED_FILE_TOKEN)
	}
	return nil
}

const UNSEALED_FILE_TOKEN = "unsealed"

func makeUnsealCmd(context *CmdContext) *cobra.Command {

	cmdArgs := app.UnsealConfigArgs{}

	cmd := &cobra.Command{
		Use:   "unseal",
		Short: "Unseal the sensitive values of a configuration with a private key",
		Long: `
	Unseal parses a config file and replaces the sealed sensitive data, such as passwords, with 
	their plaintext values.  Since asymmetric encryption is used, the private/secret key is
	required for unsealing.
		
	For example, consider the following YAML file:
	
	  mail:
	  accounts:
		- name: test1
		smtp:
		  sender_address: "example@example.com"
		  server_address: "smtp.example.com"
		  port: 465
		  username: joeuser@example.com
		  password: sealed(<encrypted string>)
		imap:
		  server_address: "imap.example.com"
		  port: 993
		  username: janeuser@example.com
		  password: sealed(<encrypted string>)
	  send_limits: []
	
	After running the unseal command, the output file will look like:
	
	mail:
	  accounts:
		- name: test1
		  smtp:
			sender_address: "example@example.com"
			server_address: "smtp.example.com"	
			port: 465
			username: joeuser@example.com
			password: it's a secret
		  imap:
			server_address: "imap.example.com"
			port: 993
			username: janeuser@example.com
			password: it's a secret
	  send_limits: []
	
	Repeated calls to unseal will ignore values that are already unsealed and only unseal any sealed
	values.
	`,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := resolveAndCheckUnsealArgs(&cmdArgs)
			if err != nil {
				return err
			}

			//once we get through validation, silence the usage
			cmd.SilenceUsage = true

			err = context.App.UnsealConfig(&cmdArgs, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// Here you will define your flags and configuration settings.

	//local flags
	cmdArgs.Input = cmd.Flags().StringP("input", "i", "", "The filename of the YAML config to be unsealed.")
	cmd.MarkFlagRequired("input")
	cmd.MarkFlagFilename("input", "yaml", "yml")

	cmdArgs.PrivateKey = cmd.Flags().StringP("privateKey", "k", "", "The filename of the private key used to seal the config.")
	cmd.MarkFlagRequired("privateKey")
	cmd.MarkFlagFilename("privateKey")

	cmdArgs.Passphrase = cmd.Flags().StringP("passphrase", "p", "", "The passphrase for the private key, if there is one.")

	cmdArgs.Output = cmd.Flags().StringP("output", "o", "", "The filename to write the output to.  If omitted, the input file path is used with '.unsealed' injected before the extension.")
	cmd.MarkFlagFilename("output")

	cmdArgs.ForceOverwrite = cmd.Flags().BoolP("forceOverwrite", "f", false, "If set, overwrite an existing file with the output.")

	return cmd

}

func resolveAndCheckUnsealArgs(args *app.UnsealConfigArgs) error {
	if *args.Output == "" {
		*args.Output = injectValueBeforeExtension(*args.Input, UNSEALED_FILE_TOKEN)
	}
	return nil
}
