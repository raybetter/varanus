package app

import (
	"fmt"
	"io"
	"varanus/internal/config"
	"varanus/internal/secrets"
	"varanus/internal/validation"
)

func (va varanusAppImpl) UnsealConfig(args *UnsealConfigArgs, outputStream io.Writer) error {

	fmt.Fprint(outputStream, args.HumanReadable())

	//load the config
	configObj, err := config.ReadConfigFromFile(*args.Input)
	if err != nil {
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
	} else {
		fmt.Fprintln(outputStream, "The config was loaded successfully.")
	}

	validationResult, err := validation.ValidateObject(configObj)
	if err != nil {
		fmt.Fprintf(outputStream, "Config validation failed: %s\n", err)
		return newApplicationError(
			"Refusing to unseal the configuration because validation had an error -- please report this as a bug: %w", err)
	}
	fmt.Fprint(outputStream, validationResult.HumanReadable())
	if validationResult.GetErrorCount() > 0 {
		return newApplicationError("Refusing to unseal unvalidated config file %s", *args.Input)
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
