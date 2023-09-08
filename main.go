package main

import (
	"fmt"
	"varanus/internal/config"

	"github.com/kr/pretty"
)

func main() {

	config, err := config.ReadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	fmt.Printf("%# v", pretty.Formatter(config))
}
