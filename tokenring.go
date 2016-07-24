package scss

import (
	"github.com/thijzert/go-scss/lexer"
)

// It's actually more of a buffer, but this sounds way cooler
type TokenRing struct {
	l      *lexer.L
	buffer []*lexer.Token
	index  int
	bts    backtrackStack
	eof    bool
}

func NewTokenRing(l *lexer.L) *TokenRing {
	rv := &TokenRing{l, make([]*lexer.Token, 0, 10), 0, newBacktrackStack(), false}
	return rv
}

func (t *TokenRing) Next() *lexer.Token {
	if t.eof {
		return nil
	} else if t.index == len(t.buffer) {
		n, _ := t.l.NextToken()
		if n == nil {
			t.eof = true
			return nil
		}
		t.buffer = append(t.buffer, n)
	}
	rv := t.buffer[t.index]
	t.index++
	return rv
}

func (t *TokenRing) Rewind() {
	t.index--
}

func (t *TokenRing) NRewind(i int) {
	t.index -= i
}

func (t *TokenRing) Mark() {
	t.bts.push(t.index)
}

func (t *TokenRing) Backtrack() {
	t.index = t.bts.pop()
}

func (t *TokenRing) Unmark() {
	t.bts.pop()
}

func (t *TokenRing) EOF() bool {
	return t.eof
}

func (t *TokenRing) Peek() *lexer.Token {
	rv := t.Next()
	t.Rewind()
	return rv
}

// A stack that keeps track of your marks
type backtrackNode struct {
	index int
	next  *backtrackNode
}

type backtrackStack struct {
	start *backtrackNode
}

func newBacktrackStack() backtrackStack {
	return backtrackStack{}
}

func (s *backtrackStack) push(index int) {
	node := &backtrackNode{index: index}
	if s.start == nil {
		s.start = node
	} else {
		node.next = s.start
		s.start = node
	}
}

func (s *backtrackStack) pop() int {
	if s.start == nil {
		return -1
	} else {
		n := s.start
		s.start = n.next
		return n.index
	}
}

func (s *backtrackStack) clear() {
	s.start = nil
}
