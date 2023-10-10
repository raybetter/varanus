package app

import (
	"fmt"
	"io"
	"varanus/internal/config"
	"varanus/internal/secrets"
	"varanus/internal/validation"
)

func (va varanusAppImpl) SealConfig(args *SealConfigArgs, outputStream io.Writer) error {

	fmt.Fprint(outputStream, args.HumanReadable())

	//load the config
	config, err := config.ReadConfigFromFile(*args.Input)
	if err != nil {
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
	} else {
		fmt.Fprintln(outputStream, "The config was loaded successfully.")
	}

	validationResult, err := validation.ValidateObject(config)
	if err != nil {
		fmt.Fprintf(outputStream, "Config validation failed: %s\n", err)
		return newApplicationError(
			"Refusing to seal the configuration because validation had an error -- please report this as a bug: %w", err)
	}
	fmt.Fprint(outputStream, validationResult.HumanReadable())
	if validationResult.GetErrorCount() > 0 {
		return newApplicationError("Refusing to seal unvalidated config file %s", *args.Input)
	}

	//create the sealer
	sealer := secrets.MakeSecretSealer()
	err = sealer.LoadPublicKeyFromFile(*args.PublicKey)
	if err != nil {
		return newApplicationError("Could not load public key from '%s': %w", *args.PublicKey, err)
	}

	//seal the config
	sealResult := sealer.SealObject(config)
	fmt.Fprint(outputStream, sealResult.HumanReadable())
	if len(sealResult.SealErrors) > 0 {
		return newApplicationError("There were errors when sealing the config.  Check the output for details.")
	}

	//write the config out
	err = config.WriteConfigToFile(*args.Output, *args.ForceOverwrite)
	if err != nil {
		return newApplicationError("Error writing output config to '%s': %w", *args.Output, err)
	}

	fmt.Fprintf(outputStream, "Seal operation succeeded.  Results written to '%s'\n", *args.Output)

	return nil
}
