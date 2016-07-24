package lexer

type runeNode struct {
	r    rune
	l, c int
	next *runeNode
}

type runeStack struct {
	start *runeNode
}

func newRuneStack() runeStack {
	return runeStack{}
}

func (s *runeStack) push(r rune, l, c int) {
	node := &runeNode{r: r, l: l, c: c}
	if s.start == nil {
		s.start = node
	} else {
		node.next = s.start
		s.start = node
	}
}

func (s *runeStack) pop() (rune, int, int) {
	if s.start == nil {
		return EOFRune, 0, 0
	} else {
		n := s.start
		s.start = n.next
		return n.r, n.l, n.c
	}
}

func (s *runeStack) clear() {
	s.start = nil
}
