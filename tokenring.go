package scss

import (
	"github.com/thijzert/go-scss/lexer"
)

// It's actually more of a stack, but this sounds way cooler
type TokenRing struct {
	l     *lexer.L
	stack []*lexer.Token
	index int
	eof   bool
}

func NewTokenRing(l *lexer.L) *TokenRing {
	rv := &TokenRing{l, make([]*lexer.Token, 0, 10), 0, false}
	return rv
}

func (t *TokenRing) Next() *lexer.Token {
	if t.eof {
		return nil
	} else if t.index == len(t.stack) {
		n, _ := t.l.NextToken()
		if n == nil {
			t.eof = true
			return nil
		}
		t.stack = append(t.stack, n)
	}
	rv := t.stack[t.index]
	t.index++
	return rv
}

func (t *TokenRing) Rewind() {
	t.index--
}

func (t *TokenRing) NRewind(i int) {
	t.index -= i
}

func (t *TokenRing) EOF() bool {
	return t.eof
}

func (t *TokenRing) Peek() *lexer.Token {
	rv := t.Next()
	t.Rewind()
	return rv
}
