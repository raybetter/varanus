package app

import (
	"fmt"
	"varanus/internal/config"
	"varanus/internal/secrets"
	"varanus/internal/validation"
)

type varanusAppImpl struct {
}

func CreateApp() VaranusApp {
	return varanusAppImpl{}
}

func newApplicationError(format string, args ...interface{}) error {
	return ApplicationError{
		theError: fmt.Errorf(format, args...),
	}
}

func (va varanusAppImpl) SealConfig(args *SealConfigArgs) error {
	fmt.Println("Sealing config")
	fmt.Println("  Input: ", *args.Input)
	fmt.Println("  PublicKey: ", *args.PublicKey)
	fmt.Println("  Output: ", *args.Output)
	fmt.Println("  ForceOverwrite: ", *args.ForceOverwrite)

	//load the config
	config, err := config.ReadConfig(*args.Input)
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
	sealResult, err := sealer.SealObject(config)
	if err != nil {
		return newApplicationError("Error while sealing the config: %w", err)
	}
	sealResult.Dump()
	if len(sealResult.SealErrors) > 0 {
		return newApplicationError("There were errors when sealing the config.  Check the output for details.")
	}

	//write the config out
	err = config.WriteConfig(*args.Output, *args.ForceOverwrite)
	if err != nil {
		return newApplicationError("Error writing output config to '%s': %w", *args.Output, err)
	}

	fmt.Printf("Seal operation succeeded.  Results written to '%s'\n", *args.Output)

	return nil
}

func (va varanusAppImpl) CheckConfig(args *CheckConfigArgs) error {
	fmt.Println("Checking config with:")
	fmt.Println("  Input: ", *args.Input)
	fmt.Println("  PrivateKey: ", *args.PrivateKey)
	fmt.Printf("  Passphrase: <redacted value of length %d>\n", len(*args.Passphrase))

	overallCheckOk := true

	//load the config
	config, err := config.ReadConfig(*args.Input)
	if err != nil {
		//return from here because we can't continue checks without a config
		return newApplicationError("Could not load config from '%s': %w", *args.Input, err)
	} else {
		fmt.Println("The config was loaded successfully.")
	}

	vp := validation.ValidationProcess{}
	config.Validate(&vp)
	err = vp.GetFinalValidationError()
	if err != nil {
		valErr, ok := err.(validation.ValidationError)
		if ok {
			fmt.Printf("There were %d validation errors:\n", len(valErr.ErrorList))
			for _, validationError := range valErr.ErrorList {
				fmt.Printf("  %s\n", validationError.Error)
			}
		} else {
			fmt.Printf("Config validation failed: %s\n", err)
		}
		// set the overallCheckOK because we can continue with checking the config even with
		// validation errors
		overallCheckOk = false
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
	sealCheckResult, err := secrets.CheckSealsOnObject(config, unsealer)
	if err != nil {
		return newApplicationError("Check operation failed unexpectedly: %w", err)
	}
	sealCheckResult.Dump()
	if len(sealCheckResult.UnsealErrors) > 0 {
		overallCheckOk = false
	}
	if unsealer == nil {
		if sealCheckResult.SealedCount > 0 {
			fmt.Println("The integrity of sealed values was not checked because no private key was provided.")
		}
	} else {
		fmt.Println("The integrity of the sealed values was verified with the private key.")
	}

	if overallCheckOk {
		fmt.Println("The configuration appears to be valid.  Congratulations!")
		return nil
	} else {
		return newApplicationError("There are some issues with the configuration.  See the output above for details.")
	}
}

func (va varanusAppImpl) UnsealConfig(args *UnsealConfigArgs) error {
	fmt.Println("Unsealing config")
	fmt.Println("  Input: ", *args.Input)
	fmt.Println("  PrivateKey: ", *args.PrivateKey)
	fmt.Printf("  Passphrase: <redacted value of length %d>\n", len(*args.Passphrase))
	fmt.Println("  Output: ", *args.Output)
	fmt.Println("  ForceOverwrite: ", *args.ForceOverwrite)

	//load the config
	config, err := config.ReadConfig(*args.Input)
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
	unsealResult, err := unsealer.UnsealObject(config)
	if err != nil {
		return newApplicationError("Error while unsealing the config: %w", err)
	}
	unsealResult.Dump()
	if len(unsealResult.UnsealErrors) > 0 {
		return newApplicationError("There were errors when unsealing the config.  Check the output for details.")
	}

	//write the config out
	err = config.WriteConfig(*args.Output, *args.ForceOverwrite)
	if err != nil {
		return newApplicationError("Error writing output config to '%s': %w", *args.Output, err)
	}

	fmt.Printf("Unseal operation succeeded.  Results written to '%s'\n", *args.Output)

	return nil
}
