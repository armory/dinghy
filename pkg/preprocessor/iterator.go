package preprocessor

type iterator struct {
	length int
	text string
	pos int
}

func newIterator(s string) *iterator {
	return &iterator{
		text: s,
		length: len(s),
		pos: 0,
	}
}

func (it *iterator) get() rune {
	if it.end() {
		return 0
	}
	return rune(it.text[it.pos])
}

func (it *iterator) end() bool {
	return it.pos >= it.length
}

func (it *iterator) slice(begin int) string {
	return it.text[begin:it.pos]
}
