package main

import (
	fcli "family-catering/cmd/cli"
	"family-catering/pkg/logger"
)

func main() {
	err := fcli.Execute()
	if err != nil {
		logger.Fatal(err, "error executing cmd app")
	}
}
