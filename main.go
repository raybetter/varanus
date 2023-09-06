package main

import (
	"fmt"
	"varanus/internal/mail"

	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	// viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
	// viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath(".")    // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	config, err := mail.LoadMailConfig(viper.GetViper())
	if err != nil {
		panic(fmt.Errorf("failed to load mail config because %s", err))
	}

	mail.PrintConfig(config)

}
