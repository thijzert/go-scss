package scss

import (
	"fmt"
	"github.com/thijzert/go-scss/lexer"
	"strings"
)

type CompileError struct {
	Message  string
	Previous error
}

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

func (p CompileError) Error() string {
	return p.Message
}

func (p CompileError) Cause() error {
	return p.Previous
}

func compileError(err string, cause error) error {
	return CompileError{err, cause}
}

func (p CompileError) String() string {
	rv := p.Message
	if p.Previous != nil {
		if perr, ok := p.Previous.(CompileError); ok {
			return rv + "\n\t" + strings.Replace(perr.String(), "\n", "\n\t", -1)
		} else if perr, ok := p.Previous.(ParseError); ok {
			return rv + "\n\t" + strings.Replace(perr.String(), "\n", "\n\t", -1)
		} else {
			return rv + "\n\t" + strings.Replace(p.Previous.Error(), "\n", "\n\t", -1)
		}
	}
	return rv
}
