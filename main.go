package main

import (
	"os"
	"varanus/cmd"
	"varanus/internal/app"
)

func main() {

	context := &cmd.CmdContext{
		App: app.CreateApp(),
	}

	command := cmd.MakeRootCmd(context)
	err := command.Execute()
	if err != nil {
		//don't need to print the error because Cobra already prints it, just set the return type
		//based on what kind of error
		_, ok := err.(app.ApplicationError)
		if ok {
			os.Exit(2)
		} else {
			os.Exit(1)
		}
	}

}
