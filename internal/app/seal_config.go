package app

import (
	"fmt"
	"io"
	"varanus/internal/config"
	"varanus/internal/secrets"
)

func (va varanusAppImpl) SealConfig(args *SealConfigArgs, outputStream io.Writer) error {

	fmt.Fprint(outputStream, args.HumanReadable())

	//load the config
	config, err := config.ReadConfigFromFile(*args.Input)
	if err != nil {
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
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
