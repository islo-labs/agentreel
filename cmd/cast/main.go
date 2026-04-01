package main

import (
	"os"

	"github.com/adamgold/agentcast/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
