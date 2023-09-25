/*
Copyright © 2023 Justin Ray

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
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
