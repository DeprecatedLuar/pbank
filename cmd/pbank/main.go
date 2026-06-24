package main

import (
	"os"

	"github.com/DeprecatedLuar/pbank/internal/orchestrator"
)

func main() {
	os.Exit(orchestrator.Run(os.Args))
}
