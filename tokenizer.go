package scss

import (
	"github.com/thijzert/go-scss/lexer"
)

const (
	WhitespaceToken lexer.TokenType = iota
	OperatorToken
	SymbolToken
	StringToken
)

func nullState(l *lexer.L) lexer.StateFunc {
	peek := l.Peek()
	if peek == lexer.EOFRune {
		return nil
	} else if peek == ' ' || peek == '\t' || peek == '\n' || peek == '\r' {
		return whitespaceState
	} else if peek == '/' {
		return commentState
	} else if peek == '"' {
		return stringState
	} else if isOperator(peek) {
		l.Next()
		l.Emit(OperatorToken)
		return nullState
	}
	return symbolState
}

func isWhitespace(r rune) bool {
	if r == ' ' {
		return true
	} else if r == '\t' {
		return true
	} else if r == '\n' {
		return true
	} else if r == '\r' {
		return true
	} else {
		return false
	}
}
func isOperator(r rune) bool {
	if r == '.' {
		return true
	} else if r == '{' {
		return true
	} else if r == '}' {
		return true
	} else if r == '>' {
		return true
	} else if r == '+' {
		return true
	} else if r == '(' {
		return true
	} else if r == ')' {
		return true
	} else if r == '[' {
		return true
	} else if r == ']' {
		return true
	} else if r == ':' {
		return true
	} else if r == ';' {
		return true
	} else if r == '\'' {
		return true
	} else if r == '"' {
		return true
	} else if r == '^' {
		return true
	} else if r == '=' {
		return true
	} else if r == ',' {
		return true
	} else if r == '*' {
		return true
	} else if r == '&' {
		return true
	} else {
		return false
	}
}

func whitespaceState(l *lexer.L) lexer.StateFunc {
	l.Take(" \t\n\r")
	l.Ignore()
	l.Emit(WhitespaceToken)

	return nullState
}

func symbolState(l *lexer.L) lexer.StateFunc {
	for {
		r := l.Next()
		if r == lexer.EOFRune {
			l.Emit(SymbolToken)
			return nil
		} else if isOperator(r) || isWhitespace(r) {
			l.Rewind()
			l.Emit(SymbolToken)
			return nullState
		}
	}
}

func commentState(l *lexer.L) lexer.StateFunc {
	peek := l.Next()
	peek = l.Peek()
	if peek == '/' {
		// Line comment
		for peek != lexer.EOFRune && peek != '\n' {
			peek = l.Next()
		}
		l.Ignore()
	} else if peek == '*' {
		// Block comment
		for {
			for peek != lexer.EOFRune && peek != '*' {
				peek = l.Next()
			}
			if peek != lexer.EOFRune {
				peek = l.Next()
				if peek == lexer.EOFRune || peek == '/' {
					l.Ignore()
					break
				}
			}
		}
	} else {
		// Well that's weird. Let's treat the slash as an operator.
		l.Emit(OperatorToken)
	}
	return nullState
}

func stringState(l *lexer.L) lexer.StateFunc {
	peek := l.Next()
	if peek != '"' {
		l.Rewind()
		return nullState
	}
	nope := false
	peek = l.Next()
	for peek != '"' || nope {
		peek = l.Next()
		nope = (peek == '\\')
	}
	l.Emit(StringToken)
	return nullState
}
