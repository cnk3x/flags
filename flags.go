package flags

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type FlagSet struct {
	*pSet
	Name        string
	Description string
	Version     string
	BuildTime   time.Time

	envKeys map[string][]string
}

type (
	pSet  = pflag.FlagSet // hidden FlagSet.FlagSet
	pFlag = pflag.Flag
)

type Option func(*FlagSet)

func SetBuildTime(buildTime time.Time) Option { return func(fs *FlagSet) { fs.BuildTime = buildTime } }
func SetVersion(version string) Option        { return func(fs *FlagSet) { fs.Version = version } }
func SetDescription(desc string) Option       { return func(fs *FlagSet) { fs.Description = desc } }

func NewSet(options ...Option) *FlagSet {
	fs := &FlagSet{}
	for _, apply := range options {
		apply(fs)
	}

	if fs.Name == "" {
		fs.Name = filepath.Base(os.Args[0])
	}

	fs.pSet = pflag.NewFlagSet(fs.Name, pflag.ExitOnError)
	return fs
}

func (fs *FlagSet) ParseFrom(args []string) (err error) {
	pflag.ErrHelp = fmt.Errorf("\nstart with %s [...OPTIONS]", filepath.Base(os.Args[0]))
	fs.SortFlags = false
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, filepath.Base(os.Args[0]))
		if fs.Version != "" {
			fmt.Fprintf(os.Stderr, " - version %s", fs.Version)
		}
		if !fs.BuildTime.IsZero() {
			fmt.Fprintf(os.Stderr, " - build %s", fs.BuildTime.In(time.Local).Format(time.RFC3339))
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "wrap system env as synology for xunlei")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		fmt.Fprintln(os.Stderr, fs.FlagUsagesWrapped(0))
	}

	reDeprecated := regexp.MustCompile(`\s*\*\*(.+)\*\*\s*`)
	fs.VisitAll(func(f *pFlag) {
		if matches := reDeprecated.FindStringSubmatch(f.Usage); len(matches) > 1 {
			f.Usage = reDeprecated.ReplaceAllString(f.Usage, "")
			f.Deprecated = matches[1]
		}

		if keys, find := fs.envKeys[f.Name]; find && len(keys) > 0 {
			f.Usage = fmt.Sprintf("%s[%s]", f.Usage, strings.Join(keys, ", "))

			if s := GetEnv(keys, ""); s != "" {
				if err := f.Value.Set(s); err != nil {
					fmt.Fprintf(os.Stderr, "WARN: set flag `%s` value `%s` from environ: %s\n", f.Name, s, err)
				}
			}
		}
	})
	return fs.pSet.Parse(os.Args[1:])
}

func (fs *FlagSet) Parse() bool { return fs.ParseFrom(os.Args[1:]) == nil }

func (fs *FlagSet) Var(v any, name, short, usage string, env ...string) {
	if len(env) > 0 {
		if usage != "" {
			usage += " "
		}
		usage = fmt.Sprintf("%s[%s]", usage, strings.Join(env, ", "))
	}

	switch x := v.(type) {
	case *time.Duration:
		fs.DurationVarP(x, name, short, *x, usage)
	case *net.IP:
		fs.IPVarP(x, name, short, *x, usage)
	case *net.IPNet:
		fs.IPNetVarP(x, name, short, *x, usage)
	case *string:
		fs.StringVarP(x, name, short, *x, usage)
	case *int:
		fs.IntVarP(x, name, short, *x, usage)
	case *int8:
		fs.Int8VarP(x, name, short, *x, usage)
	case *int16:
		fs.Int16VarP(x, name, short, *x, usage)
	case *int32:
		fs.Int32VarP(x, name, short, *x, usage)
	case *int64:
		fs.Int64VarP(x, name, short, *x, usage)
	case *uint:
		fs.UintVarP(x, name, short, *x, usage)
	case *uint8:
		fs.Uint8VarP(x, name, short, *x, usage)
	case *uint16:
		fs.Uint16VarP(x, name, short, *x, usage)
	case *uint32:
		fs.Uint32VarP(x, name, short, *x, usage)
	case *uint64:
		fs.Uint64VarP(x, name, short, *x, usage)
	case *float32:
		fs.Float32VarP(x, name, short, *x, usage)
	case *float64:
		fs.Float64VarP(x, name, short, *x, usage)
	case *bool:
		fs.BoolVarP(x, name, short, *x, usage)
	case *[]time.Duration:
		fs.DurationSliceVarP(x, name, short, *x, usage)
	case *[]net.IP:
		fs.IPSliceVarP(x, name, short, *x, usage)
	case *[]net.IPNet:
		fs.IPNetSliceVarP(x, name, short, *x, usage)
	case *[]string:
		fs.StringSliceVarP(x, name, short, *x, usage)
	case *[]int:
		fs.IntSliceVarP(x, name, short, *x, usage)
	case *[]int32:
		fs.Int32SliceVarP(x, name, short, *x, usage)
	case *[]int64:
		fs.Int64SliceVarP(x, name, short, *x, usage)
	case *[]uint:
		fs.UintSliceVarP(x, name, short, *x, usage)
	case *[]float32:
		fs.Float32SliceVarP(x, name, short, *x, usage)
	case *[]float64:
		fs.Float64SliceVarP(x, name, short, *x, usage)
	case *[]bool:
		fs.BoolSliceVarP(x, name, short, *x, usage)
	default:
		panic(fmt.Errorf("%s type %v(%T) not support", name, x, x))
	}

	if len(env) > 0 {
		fs.envKeys[name] = env
	}
}

func GetEnv(keys []string, def string) (s string) {
	if len(keys) > 0 {
		for _, e := range keys {
			if s = os.Getenv(e); s != "" {
				return
			}
		}
	}
	return def
}
