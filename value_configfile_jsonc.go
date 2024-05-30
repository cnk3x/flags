package flags

const (
	jc_ESCAPE   = 92
	jc_QUOTE    = 34
	jc_SPACE    = 32
	jc_TAB      = 9
	jc_NEWLINE  = 10
	jc_ASTERISK = 42
	jc_SLASH    = 47
	jc_HASH     = 35
)

func jcTranslate(s []byte) []byte {
	if len(s) <= 2 {
		return s
	}

	var (
		i       int
		quote   bool
		escaped bool
	)
	j := make([]byte, len(s))
	comment := &jcCommentData{}
	for _, ch := range s {
		if ch == jc_ESCAPE || escaped {
			if !comment.startted {
				j[i] = ch
				i++
			}
			escaped = !escaped
			continue
		}
		if ch == jc_QUOTE && !comment.startted {
			quote = !quote
		}
		if (ch == jc_SPACE || ch == jc_TAB) && !quote {
			continue
		}
		if ch == jc_NEWLINE {
			if comment.isSingleLined {
				comment.stop()
			}
			continue
		}
		if quote && !comment.startted {
			j[i] = ch
			i++
			continue
		}
		if comment.startted {
			if ch == jc_ASTERISK && !comment.isSingleLined {
				comment.canEnd = true
				continue
			}
			if comment.canEnd && ch == jc_SLASH && !comment.isSingleLined {
				comment.stop()
				continue
			}
			comment.canEnd = false
			continue
		}
		if comment.canStart && (ch == jc_ASTERISK || ch == jc_SLASH) {
			comment.start(ch)
			continue
		}
		if ch == jc_SLASH {
			comment.canStart = true
			continue
		}
		if ch == jc_HASH {
			comment.start(ch)
			continue
		}
		j[i] = ch
		i++
	}
	return j[:i]
}

type jcCommentData struct {
	canStart      bool
	canEnd        bool
	startted      bool
	isSingleLined bool
}

func (c *jcCommentData) stop() {
	c.startted = false
	c.canStart = false
}

func (c *jcCommentData) start(ch byte) {
	c.startted = true
	c.isSingleLined = ch == jc_SLASH || ch == jc_HASH
}
