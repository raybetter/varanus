package app

import (
	"fmt"
	"io"
	"varanus/internal/config"
	"varanus/internal/secrets"
)

func (va varanusAppImpl) UnsealConfig(args *UnsealConfigArgs, outputStream io.Writer) error {

	fmt.Fprint(outputStream, args.HumanReadable())

	//load the config
	configObj, err := config.ReadConfigFromFile(*args.Input)
	if err != nil {
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
	}

	//create the unsealer
	unsealer := secrets.MakeSecretUnsealer()
	//TODO implement the ways this can handle passphrases work like openssl, like "env:, pass:, file:"
	err = unsealer.LoadPrivateKeyFromFile(*args.PrivateKey, *args.Passphrase)
	if err != nil {
		return newApplicationError("Could not load private key from '%s': %w", *args.PrivateKey, err)
	}

	//unseal the config
	unsealResult := unsealer.UnsealObject(configObj)
	fmt.Fprint(outputStream, unsealResult.HumanReadable())
	if len(unsealResult.UnsealErrors) > 0 {
		return newApplicationError("There were errors when unsealing the config.  Check the output for details.")
	}

	//write the config out
	err = configObj.WriteConfigToFile(*args.Output, *args.ForceOverwrite)
	if err != nil {
		return newApplicationError("Error writing output config to '%s': %w", *args.Output, err)
	}

	fmt.Fprintf(outputStream, "Unseal operation succeeded.  Results written to '%s'\n", *args.Output)

	return nil
}
