package flags

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type (
	FlagSet = pflag.FlagSet
	Flag    = pflag.Flag
	Value   = pflag.Value
)

var Default = func() *FlagSet { return pflag.CommandLine }

var version string

// 设置或者获取版本号， updateVer参数的第一个不为空的值将设置到版本号中，返回最终版本号
func Version(updateVer ...string) string {
	for _, ver := range updateVer {
		if ver != "" {
			version = ver
		}
	}
	return version
}

func name() string { return filepath.Base(os.Args[0]) }

func ParseFlags(set *FlagSet, args []string) (err error) {
	name, out := name(), os.Stderr

	pflag.ErrHelp = fmt.Errorf("use %s [...OPTIONS] to start", name)

	set.Init(name, pflag.ContinueOnError)
	set.SetOutput(out)
	set.SortFlags = false

	set.Usage = func() {
		fmt.Fprintf(out, "%s", name)
		if version != "" {
			fmt.Fprintf(out, " -- version %s", version)
		}
		fmt.Fprintf(out, "\n\n")
		fmt.Fprintf(out, "USAGE:\n")
		fmt.Fprintf(out, "      %s [...OPTIONS]\n\n", name)
		fmt.Fprintf(out, "OPTIONS:\n")
		fmt.Fprintln(out, set.FlagUsagesWrapped(0))
		fmt.Fprintln(out)
	}

	if set == Default() {
		pflag.Usage = set.Usage
	}

	if version != "" && set.Lookup("version") == nil {
		var shorthand string
		if set.Lookup("v") == nil {
			shorthand = "v"
		}
		if shorthand == "" && set.Lookup("V") == nil {
			shorthand = "V"
		}
		set.BoolP("version", shorthand, false, "显示版本号")
	}

	if err = set.Parse(args); err != nil {
		return
	}

	if ver, _ := set.GetBool("version"); ver {
		fmt.Fprintf(out, "%s", name)
		if version != "" {
			fmt.Fprintf(out, " -- version %s", version)
		}
		os.Exit(0)
	}

	return
}

func Parse() {
	if err := ParseFlags(Default(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func flagSet(flags []*FlagSet) *FlagSet {
	for _, f := range flags {
		if f != nil {
			return f
		}
	}
	return Default()
}
