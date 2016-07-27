package scss

import (
	"github.com/thijzert/go-scss/lexer"
)

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
	tok.Mark()
	rv.Rules = make([]Rule, 0)

	peek := tok.Ignore(WhitespaceToken)
	tok.Rewind()

	var rule Rule
	for peek != nil {
		if peek.Type == OperatorToken && (peek.Value == "@" || peek.Value == "$") {
			// TODO: handle @import, $macro, @mixin...
			if peek.Value == "@" {
				err = parseError("@-directives are not implemented", nil, peek)
				tok.Backtrack()
				return
			} else if peek.Value == "$" {
				err = parseError("Macros are not implemented", nil, peek)
				tok.Backtrack()
				return
			}
		}

		rule, err = parseRule(tok)

		if err != nil {
			err = parseError("Error parsing rule", err, peek)
			tok.Backtrack()
			return
		}

		rv.Rules = append(rv.Rules, rule)

		peek = tok.Ignore(WhitespaceToken)
		if peek != nil {
			tok.Rewind()
		}
	}

	tok.Unmark()
	return
}

func parseRule(tok *TokenRing) (rv Rule, err error) {
	tok.Mark()
	rv.Selector, err = parseSelector(tok)
	if err != nil {
		err = parseError("Error parsing selector list", err, tok.Peek())
		tok.Backtrack()
		return
	}

	rv.Scope, err = parseScope(tok)
	if err != nil {
		err = parseError("Error parsing scope", err, tok.Peek())
		tok.Backtrack()
		return
	}
	tok.Unmark()
	return
}

func parseScope(tok *TokenRing) (rv Scope, err error) {
	tok.Mark()

	peek := tok.Next()
	if peek == nil || peek.Type != OperatorToken || peek.Value != "{" {
		err = parseError("Expected: '{'", nil, peek)
		tok.Backtrack()
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
				tok.Backtrack()
				return
			}
			rv.Subrules = append(rv.Subrules, rule)
		}
		peek = tok.Ignore(WhitespaceToken)
		if peek != nil {
			tok.Rewind()
		}
	}

	peek = tok.Next()
	if peek == nil || peek.Type != OperatorToken || peek.Value != "}" {
		err = parseError("Expected: '}'", nil, peek)
		tok.Backtrack()
		return
	}

	tok.Unmark()
	return
}

func parseProperty(tok *TokenRing) (rv Property, err error) {
	tok.Mark()

	peek := tok.Ignore(WhitespaceToken)
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.Backtrack()
		return
	}
	if peek.Type != SymbolToken {
		err = parseError("expected symbol", nil, peek)
		tok.Backtrack()
		return
	}

	rv.Key = peek.Value

	peek = tok.Ignore(WhitespaceToken)
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.Backtrack()
		return
	}
	if peek.Type != OperatorToken || peek.Value != ":" {
		err = parseError("expected ':'", nil, peek)
		tok.Backtrack()
		return
	}

	peek = tok.Next()
	if peek == nil {
		err = parseError("unexpected EOF", nil, peek)
		tok.Backtrack()
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
			tok.Unmark()
			return
		} else {
			rv.Value = rv.Value + peek.Value
		}
		peek = tok.Next()
	}

	tok.Unmark()
	return
}
