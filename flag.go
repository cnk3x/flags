package flags

import (
	"flag"
)

var (
	CommandLine = flag.CommandLine
	NewFlagSet  = flag.NewFlagSet
)

type FlagSet = flag.FlagSet
