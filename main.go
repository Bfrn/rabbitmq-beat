package main

import (
	"os"

	"hummer/rabbitmq-beat/cmd"

	_ "hummer/rabbitmq-beat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
