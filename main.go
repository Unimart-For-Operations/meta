package main

import (
	"os"

	"github.com/idpbuilder/meta/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
