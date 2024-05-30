package flags

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

var FmtPrintln = func(s string) { fmt.Println("  --" + s) }
var FmtPrintf = func(s string, args ...any) { fmt.Printf(s+"\n", args...) }

type TestStruct struct {
	Str      string   `flag:"str,STR"`
	Strs     []string `flag:"strs,YES,s"`
	Int64p   *int64
	Durpsp   *[]*time.Duration
	Durps    []*time.Duration
	Durs     []time.Duration
	Int64s   []int64
	Uint64ps []*uint64
	Uintsp   *[]uint
	Time     time.Time
	Times    []time.Time
	Timep    *time.Time
	Hello    bool
}

func TestFlag(t *testing.T) {
	Version("1.0.0")
	time.Local = time.FixedZone("CST", 8*3600)

	os.Setenv("STR", ":9981")
	os.Args = []string{
		os.Args[0],
		"--str", "a",
		"--str", "b",
		"--strs", "b", "--strs", "c",
		"--int64p", "13",
		"--durpsp", "1s", "--durpsp", "1h", "--durpsp", "1d",
		"--durps", "1s", "--durps", "1h", "--durps", "1d",
		// "--durs", "1s", "--durs", "1h",
		// "--durs", "1d",
		"--int64s", "1", "--int64s", "2", "--int64s", "3",
		"--uint64ps", "1", "--uint64ps", "2", "--uint64ps", "3",
		"--uintsp", "1", "--uintsp", "2", "--uintsp", "3",
		"--time", "2021-01-01",
		"--times", "2021-01-01", "--times", "2021-01-02", "--times", "2021-01-03 11:12",
		"--timep", "2022/01/01 11:12",
		"--hello",
		// "-h",
	}

	var cfg TestStruct

	cfg.Durs = []time.Duration{1 * time.Second, 1 * time.Hour}

	StructBind(&cfg)
	Parse()

	StructPrint(cfg, FmtPrintln)

	args := StructToArgs(cfg)

	fmt.Printf("args:  \"%s\"\n", strings.Join(args, `", "`))

	pflag.Usage()
}

const FMT = "| %-15s | %-17s | %-5s | %-5s | %-5s | %-5s | %-5s | %-5s |"

func TestReflectStruct(t *testing.T) {
	s := TestStruct{Times: []time.Time{time.Now()}, Str: "1", Int64s: []int64{}}
	sv := rVal(&s)
	st := sv.Type().Elem()
	sv = sv.Elem()

	metaPrint := metaPrinter(FmtPrintf)

	metaPrint("Main", sv, st)
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		metaPrint(ft.Name, sv.Field(i), ft.Type)
	}
}

func TestValue(t *testing.T) {
	metaPrint := metaPrinter(FmtPrintf)

	var (
		conf      string = "hello"
		vDuration time.Duration
		vInt64    int64 = 3
		vInt      int
		vIntP     **int
		vInts     []int
		vIntsP    *[]int
		vIntPs    []*int
		vIntPsP   *[]*int
	)

	rSets(rVal(&vDuration), "24d1h2s")
	rSets(rVal(&vInt64), "100")
	rSets(rVal(&vInt), "101")
	rSets(rVal(&vIntP), "102")
	rSets(rVal(&vInts), "103")
	rSets(rVal(&vInts), "104")
	rSets(rVal(&vIntsP), "105")
	rSets(rVal(&vIntsP), "106")

	metaPrint("conf", rVal(&conf), nil)
	metaPrint("vDuration", rVal(&vDuration), nil)
	metaPrint("vInt64", rVal(&vInt64), nil)
	metaPrint("vInt", rVal(&vInt), nil)
	metaPrint("vIntP", rVal(&vIntP), nil)
	metaPrint("vInts", rVal(&vInts), nil)
	metaPrint("vIntsP", rVal(&vIntsP), nil)
	metaPrint("vIntPs", rVal(&vIntPs), nil)
	metaPrint("vIntPsP", rVal(&vIntPsP), nil)
}

func metaPrinter(printf func(string, ...any)) func(name string, rv reflect.Value, rt reflect.Type) {
	headers := []string{
		"name      ",
		"type             ",
		"ptr",
		"set",
		"addr",
		"valid",
		"nil",
		"zero",
	}

	headerPrint := sync.OnceFunc(func() {
		printf("| %s |", strings.Join(headers, " | "))
		printf("|%s|-------", strings.Join(sliceMap(headers, func(s string) string { return strings.Repeat("-", len(s)+2) }), "|"))
	})

	return func(name string, rv reflect.Value, rt reflect.Type) {
		headerPrint()

		if rt == nil {
			rt = rv.Type()
		}
		printf(
			"|"+strings.Repeat(" %-*s |", len(headers))+" %s",
			len(headers[0]),
			name,
			len(headers[1]),
			rt.String(),
			len(headers[2]),
			sBool(isPtr(rv)),
			len(headers[3]),
			sBool(rv.CanSet()),
			len(headers[4]),
			sBool(rv.CanAddr()),
			len(headers[5]),
			sBool(rv.IsValid()),
			len(headers[6]),
			sBool(isNil(rv)),
			len(headers[7]),
			sBool(rv.IsZero()),
			rGets(rv),
		)
	}
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return !v.IsValid()
	}
}

func isPtr(v reflect.Value) bool {
	return v.Kind() == reflect.Pointer
}

func sBool(b bool) string {
	if b {
		return " √"
	}
	return " \u00d7"
	// return "×"
}

func sliceMap[S any, R any](s []S, f func(S) R) []R {
	r := make([]R, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
