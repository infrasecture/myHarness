package main

import (
	"os"

	"github.com/infrasecture/myHarness/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
