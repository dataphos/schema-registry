package main

import (
	"flag"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitorctl"
)

func main() {
	configFile := flag.String("f", "", "toml file containing configuration of working environment")
	flag.Parse()

	janitorctl.RunCentralConsumer(*configFile)
}
