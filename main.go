package main

import (
	"os"

	"github.com/Unimart-For-Operations/meta/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
