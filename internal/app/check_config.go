package app

import (
	"fmt"
	"io"
	"varanus/internal/config"
	"varanus/internal/secrets"
	"varanus/internal/validation"
)

func (va varanusAppImpl) CheckConfig(args *CheckConfigArgs, outputStream io.Writer) error {

	fmt.Fprint(outputStream, args.HumanReadable())

	overallCheckOk := true

	//load the config
	config, err := config.ReadConfigFromFile(*args.Input)
	if err != nil {
		//return from here because we can't continue checks without a config
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
	} else {
		fmt.Fprintln(outputStream, "The config was loaded successfully.")
	}

	validationResult, err := validation.ValidateObject(config)
	if err != nil {
		fmt.Fprintf(outputStream, "Config validation failed: %s\n", err)
		return newApplicationError(
			"Checking did not complete because the configuration validation had an error -- please report this as a bug: %w", err)
	} else {
		fmt.Fprint(outputStream, validationResult.HumanReadable())
		if validationResult.GetErrorCount() > 0 {
			overallCheckOk = false
		}
	}

	var unsealer secrets.SecretUnsealer

	if len(*args.PrivateKey) > 0 {
		//create the unsealer
		unsealer = secrets.MakeSecretUnsealer()

		//TODO implement the ways this can handle passphrases work like openssl, like "env:, pass:, file:"
		err = unsealer.LoadPrivateKeyFromFile(*args.PrivateKey, *args.Passphrase)
		if err != nil {
			//fail if we can't load a private key because we can't do the check the user intended
			return newApplicationError("Could not load private key from '%s': %w", *args.PrivateKey, err)
		}
	}

	//we can check the seals even if the unsealer is nil
	sealCheckResult := secrets.CheckSealsOnObject(config, unsealer)
	//print context based on key presence
	if unsealer == nil {
		if sealCheckResult.SealedCount > 0 {
			fmt.Fprintln(outputStream, "The integrity of sealed values was not checked because no private key was provided.")
		}
	} else {
		if len(sealCheckResult.UnsealErrors) > 0 {
			fmt.Fprintln(outputStream, "The integrity of the sealed values was checked with the private key, but there were some errors.")
		} else {
			fmt.Fprintln(outputStream, "The integrity of the sealed values was verified with the private key.")
		}
	}
	//print seal check status, including seal errors if any
	fmt.Fprint(outputStream, sealCheckResult.HumanReadable())
	if len(sealCheckResult.UnsealErrors) > 0 {
		overallCheckOk = false
	}

	if overallCheckOk {
		fmt.Fprintln(outputStream, "The configuration appears to be valid.  Congratulations!")
		return nil
	} else {
		return newApplicationError("There are some issues with the configuration.  See the output above for details.")
	}
}
