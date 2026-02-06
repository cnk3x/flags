package flags

type Item struct {
	Name                string
	Shorthand           string
	Usage               string
	NoOptDefVal         string
	Deprecated          string
	Hidden              bool
	ShorthandDeprecated string
	Annotations         map[string][]string
	Env                 []string
}

type ItemOption func(item *Item)

func Shorthand(s string, deprecated ...string) ItemOption {
	return func(it *Item) {
		it.Shorthand = s
		if len(deprecated) > 0 {
			it.ShorthandDeprecated = deprecated[0]
		}
	}
}

func Usage(s string) ItemOption      { return func(it *Item) { it.Usage = s } }
func Deprecated(s string) ItemOption { return func(it *Item) { it.Deprecated = s } }
func Hidden() ItemOption             { return func(it *Item) { it.Hidden = true } }

func Env(s ...string) ItemOption    { return func(it *Item) { it.Env = append(it.Env, s...) } }
func EnvSet(s ...string) ItemOption { return func(it *Item) { it.Env = s } }

func Annotation(k, v string) ItemOption {
	return func(it *Item) {
		if it.Annotations == nil {
			it.Annotations = make(map[string][]string)
		}
		it.Annotations[k] = append(it.Annotations[k], v)
	}
}
