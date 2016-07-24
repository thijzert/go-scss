package scss

import (
	"fmt"
	_ "github.com/pkg/errors"
	"github.com/thijzert/go-scss/lexer"
	"strings"
)

type ParseError struct {
	Message   string
	Previous  error
	LastToken *lexer.Token
}

func (p ParseError) Error() string {
	return p.Message
}

func (p ParseError) Cause() error {
	return p.Previous
}

func parseError(err string, cause error, lastToken *lexer.Token) error {
	return ParseError{err, cause, lastToken}
}

func (p ParseError) String() string {
	rv := p.Message
	if p.LastToken != nil {
		rv = fmt.Sprintf("%s -- at line %d c %d", p.Message, p.LastToken.Line, p.LastToken.Column)
	}
	if p.Previous != nil {
		if perr, ok := p.Previous.(ParseError); ok {
			return rv + "\n\t" + strings.Replace(perr.String(), "\n", "\n\t", -1)
		} else {
			return rv + "\n\t" + strings.Replace(p.Previous.Error(), "\n", "\n\t", -1)
		}
	}
	return rv
}

type Selector []*lexer.Token
type Property struct {
	Key, Value string
}
type Scope struct {
	Properties []Property
	Subrules   []Rule
}
type Rule struct {
	Selector Selector
	Scope    Scope
}
type IR struct {
	Rules []Rule
}

func Parse(src string) (rv IR, err error) {
	l := lexer.New(src, nullState)
	l.Start()
	tok := NewTokenRing(l)
	rv, err = parseIR(tok)
	return
}

func parseIR(tok *TokenRing) (rv IR, err error) {
	rv.Rules = make([]Rule, 0)
	peek := tok.Next()
	for peek != nil && peek.Type == WhitespaceToken {
		peek = tok.Next()
	}
	tok.Rewind()

	var rule Rule
	for peek != nil {
		if peek.Type == OperatorToken && (peek.Value == "@" || peek.Value == "$") {
			// TODO: handle @import, $macro, @mixin...
			if peek.Value == "@" {
				err = parseError("@-directives are not implemented", nil, peek)
				return
			} else if peek.Value == "$" {
				err = parseError("Macros are not implemented", nil, peek)
				return
			}
		}

		rule, err = parseRule(tok)

		if err != nil {
			err = parseError("Error parsing rule", err, peek)
			return
		}

		rv.Rules = append(rv.Rules, rule)

		peek = tok.Next()
		for peek != nil && peek.Type == WhitespaceToken {
			peek = tok.Next()
		}
		if peek != nil {
			tok.Rewind()
		}
	}

	return
}

func parseRule(tok *TokenRing) (rv Rule, err error) {
	rv.Selector, err = parseSelector(tok)
	if err != nil {
		err = parseError("Error parsing selector list", err, tok.Peek())
		return
	}

	rv.Scope, err = parseScope(tok)
	if err != nil {
		err = parseError("Error parsing scope", err, tok.Peek())
		return
	}
	return
}

func parseSelector(tok *TokenRing) (rv Selector, err error) {
	c := tok.Next()
	for c != nil && c.Type == WhitespaceToken {
		c = tok.Next()
	}
	for c != nil {
		if c.Type == OperatorToken && c.Value == "{" {
			tok.Rewind()
			if len(rv) == 0 {
				err = parseError("empty selector", nil, c)
			}
			return
		} else if c.Type == OperatorToken && c.Value == "}" {
			tok.Rewind()
			if len(rv) == 0 {
				err = parseError("Unexpected '}'", nil, c)
			}
			return
		}
		rv = append(rv, c)
		c = tok.Next()
	}
	return
}

func parseScope(tok *TokenRing) (rv Scope, err error) {
	peek := tok.Next()

	if peek == nil || peek.Type != OperatorToken || peek.Value != "{" {
		err = parseError("Expected: '{'", nil, peek)
		tok.Rewind()
		return
	}
	peek = tok.Next()

	var rule Rule
	var prop Property
	for peek != nil && (peek.Type != OperatorToken || peek.Value != "}") {
		prop, err = parseProperty(tok)
		if err == nil {
			rv.Properties = append(rv.Properties, prop)
		} else {
			rule, err = parseRule(tok)
			if err != nil {
				err = parseError("Error parsing scope", err, peek)
				return
			}
			rv.Subrules = append(rv.Subrules, rule)
		}
		peek = tok.Next()
		for peek != nil && peek.Type == WhitespaceToken {
			peek = tok.Next()
		}
		if peek != nil {
			tok.Rewind()
		}
	}

	if peek == nil || peek.Type != OperatorToken || peek.Value != "}" {
		err = parseError("Expected: '}'", nil, peek)
		return
	}
	tok.Next()

	return
}

func parseProperty(tok *TokenRing) (rv Property, err error) {

	b := 1
	peek := tok.Next()
	for peek != nil && peek.Type == WhitespaceToken {
		b += 1
		peek = tok.Next()
	}
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.NRewind(b - 1)
		return
	}
	if peek.Type != SymbolToken {
		err = parseError("expected symbol", nil, peek)
		tok.NRewind(b)
		return
	}

	rv.Key = peek.Value

	b += 1
	peek = tok.Next()
	for peek != nil && peek.Type == WhitespaceToken {
		b += 1
		peek = tok.Next()
	}
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.NRewind(b - 1)
		return
	}
	if peek.Type != OperatorToken || peek.Value != ":" {
		err = parseError("expected ':'", nil, peek)
		tok.NRewind(b)
		return
	}

	b += 1
	peek = tok.Next()
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.NRewind(b - 1)
		return
	}

	rv.Value = ""
	for peek != nil {
		if peek.Type == WhitespaceToken {
			if rv.Value != "" {
				rv.Value = rv.Value + " "
			}
		} else if peek.Type == OperatorToken && (peek.Value == ";" || peek.Value == "}") {
			if peek.Value == "}" {
				tok.Rewind()
			}
			return
		} else {
			rv.Value = rv.Value + peek.Value
		}
		b += 1
		peek = tok.Next()
	}

	return
}
