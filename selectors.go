package scss

type selectorNodeType int

const (
	stImplicitAmp selectorNodeType = iota
	stExplicitAmp
	stCompoundBoth
	stCompoundEither
	stCompoundDescendant
	stCompoundDirectDescendant
	stCompoundNextSibling
	stTag
	stClass
	stPseudoclass
	stFunctionClass
	stAttribute
)

type Selector interface {
	Type() selectorNodeType
	Evaluate() string
}

func parseSelector(tok *TokenRing) (rv Selector, err error) {
	return realParseSelector(tok, &sImplicitAmp{})
}

func realParseSelector(tok *TokenRing, left Selector) (rv Selector, err error) {
	tok.Mark()
	peek := tok.Peek()
	whitespaceFound := peek != nil && peek.Type == WhitespaceToken

	committed := false
	compType := stCompoundDescendant
	if !whitespaceFound && left.Type() != stImplicitAmp {
		compType = stCompoundBoth
	}

	peek = tok.Ignore(WhitespaceToken)
	tok.Rewind()
	if peek == nil {
		err = parseError("Unexpected EOF", nil, nil)
		tok.Backtrack()
		return
	} else if peek.Type == OperatorToken {
		if peek.Value == ">" {
			committed = true
			tok.Next()
			compType = stCompoundDirectDescendant
		} else if peek.Value == "+" {
			committed = true
			tok.Next()
			compType = stCompoundNextSibling
		}
	}

	var right Selector
	right, err = parseSimpleSelector(tok)
	if err != nil || right == nil {
		if committed {
			err = parseError("Unexpected end of selector after '"+peek.Value+"'", err, peek)
			tok.Backtrack()
			return
		} else {
			return left, nil
		}
	}

	var farright Selector
	farright, err = realParseSelector(tok, right)
	if err == nil {
		right = farright
	}
	err = nil

	if compType == stCompoundDescendant {
		rv = &sCompoundDescendant{left, right}
	} else {
		err = parseError("unknown compound type", nil, peek)
	}

	if err == nil {
		tok.Unmark()
	} else {
		tok.Backtrack()
	}
	return
}

func parseSimpleSelector(tok *TokenRing) (rv Selector, err error) {
	tok.Mark()
	peek := tok.Ignore(WhitespaceToken)

	if peek.Type == SymbolToken {
		// TODO: Check against list of allowed tag names
		// FIXME: Does such a list exist?
		rv = &sTag{peek.Value}
		tok.Unmark()
	} else if peek.Type == OperatorToken && peek.Value == "." {
		// Class!
		peek = tok.Next()
		if peek == nil || peek.Type != SymbolToken {
			err = parseError("Expected symbol", nil, peek)
		} else {
			rv = &sClass{peek.Value}
		}
	} else {
		err = parseError("expected symbol or operator", nil, peek)
	}

	if err == nil {
		tok.Unmark()
	} else {
		tok.Backtrack()
	}
	return
}

type sImplicitAmp struct{}

func (s *sImplicitAmp) Type() selectorNodeType {
	return stImplicitAmp
}
func (s *sImplicitAmp) Evaluate() string {
	return ""
}

type sTag struct {
	TagName string
}

func (s *sTag) Type() selectorNodeType {
	return stTag
}
func (s *sTag) Evaluate() string {
	return s.TagName
}

type sClass struct {
	ClassName string
}

func (s *sClass) Type() selectorNodeType {
	return stClass
}
func (s *sClass) Evaluate() string {
	return "." + s.ClassName
}

type sCompoundDescendant struct {
	A, B Selector
}

func (s *sCompoundDescendant) Type() selectorNodeType {
	return stCompoundDescendant
}
func (s *sCompoundDescendant) Evaluate() string {
	return s.A.Evaluate() + " " + s.B.Evaluate()
}
