package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitorctl"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	configFile := flag.String("f", "", "toml file containing configuration of working environment")
	flag.Parse()

	janitorctl.RunCentralConsumer(*configFile)
}
