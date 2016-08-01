package scss

import (
	"fmt"
)

type selectorNodeType int

const (
	stImplicitAmp selectorNodeType = iota
	stExplicitAmp
	stCompoundBoth
	stCompoundEither
	stCompoundDescendant
	stCompoundDirectDescendant
	stCompoundNextSibling
	stID
	stTag
	stClass
	stPseudoclass
	stFunctionClass
	stAttribute
)

type Selector interface {
	Type() selectorNodeType
	Evaluate() string
	Clone() Selector
}

type lefter interface {
	Left() Selector
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
		} else if peek.Value == "," {
			if left.Type() == stImplicitAmp {
				err = parseError("unexpected ','", nil, peek)
				tok.Backtrack()
				return
			}
			committed = true
			tok.Next()
			compType = stCompoundEither
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

	if compType != stCompoundEither {
		if rr, ok := right.(*sEither); ok {
			rrv := &sEither{make([]Selector, len(rr.Terms))}
			for i, t := range rr.Terms {
				rrv.Terms[i] = &sCompound{compType, left.Clone(), t}
			}
			rv = rrv
		} else {
			rv = &sCompound{compType, left, right}
		}
	} else {
		rrv := &sEither{}
		if rr, ok := right.(*sEither); ok {
			// [.a, [.b, .c]] -> [.a, .b, .c]
			rrv.Terms = make([]Selector, len(rr.Terms)+1)
			rrv.Terms[0] = left
			copy(rrv.Terms[1:], rr.Terms)
		} else {
			rrv.Terms = make([]Selector, 2)
			rrv.Terms[0] = left
			rrv.Terms[1] = right
		}
		rv = rrv
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
	} else if peek.Type == OperatorToken {
		if peek.Value == "." {
			// Class!
			peek = tok.Next()
			if peek == nil || peek.Type != SymbolToken {
				err = parseError("Expected symbol", nil, peek)
			} else {
				rv = &sClass{peek.Value}
			}
		} else if peek.Value == "#" {
			// ID
			peek = tok.Next()
			if peek == nil || peek.Type != SymbolToken {
				err = parseError("Expected symbol", nil, peek)
			} else {
				rv = &sID{peek.Value}
			}
		} else {
			err = parseError("unexpected operator '"+peek.Value+"'", nil, peek)
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
func (s *sImplicitAmp) Clone() Selector {
	return &sImplicitAmp{}
}

type sID struct {
	ID string
}

func (s *sID) Type() selectorNodeType {
	return stID
}
func (s *sID) Evaluate() string {
	return s.ID
}
func (s *sID) Clone() Selector {
	return &sID{s.ID}
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
func (s *sTag) Clone() Selector {
	return &sTag{s.TagName}
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
func (s *sClass) Clone() Selector {
	return &sClass{s.ClassName}
}

type sCompound struct {
	CompoundType selectorNodeType
	A, B         Selector
}

func (s *sCompound) Type() selectorNodeType {
	return s.CompoundType
}
func (s *sCompound) Evaluate() string {
	if s.CompoundType == stCompoundDirectDescendant {
		return s.A.Evaluate() + ">" + s.B.Evaluate()
	} else if s.CompoundType == stCompoundDescendant {
		return s.A.Evaluate() + " " + s.B.Evaluate()
	} else if s.CompoundType == stCompoundNextSibling {
		return s.A.Evaluate() + "+" + s.B.Evaluate()
	} else if s.CompoundType == stCompoundBoth {
		return s.A.Evaluate() + s.B.Evaluate()
	} else {
		// FIXME: Detect this error in an earlier stage, and pass it through appropriate channels
		return s.A.Evaluate() + "?" + s.B.Evaluate()
	}
}
func (s *sCompound) Clone() Selector {
	return &sCompound{s.CompoundType, s.A.Clone(), s.B.Clone()}
}
func (s *sCompound) Left() Selector {
	return s.A
}

type sEither struct {
	Terms []Selector
}

func (s *sEither) Type() selectorNodeType {
	return stCompoundEither
}
func (s *sEither) Evaluate() string {
	rv := ""
	for i, ss := range s.Terms {
		if i > 0 {
			rv += ","
		}
		rv += ss.Evaluate()
	}
	return rv
}
func (s *sEither) Clone() Selector {
	rv := &sEither{make([]Selector, len(s.Terms))}
	for i, ss := range s.Terms {
		rv.Terms[i] = ss.Clone()
	}
	return rv
}

// Compose two selectors into one
func composeSelectors(top, bottom Selector) (Selector, error) {
	if tcmp, ok := top.(*sEither); ok {
		var err error
		if bcmp, ok := bottom.(*sEither); ok {
			lb := len(bcmp.Terms)
			// [[.a, .b] [.c, .d]] -> [[.a .c],[.a .d],[.b .c],[.b .d]]
			rv := &sEither{make([]Selector, lb*len(tcmp.Terms))}
			for i, t := range tcmp.Terms {
				for j, b := range bcmp.Terms {
					rv.Terms[lb*i+j], err = composeSelectors(t, b)
					if err != nil {
						return nil, err
					}
				}
			}
			return rv, nil
		} else {
			// [[.a, .b] .c] -> [[.a .c],[.b .c]]
			rv := &sEither{make([]Selector, len(tcmp.Terms))}
			for i, t := range tcmp.Terms {
				rv.Terms[i], err = composeSelectors(t, bottom)
				if err != nil {
					return nil, err
				}
			}
			return rv, nil
		}
	} else if bcmp, ok := bottom.(*sEither); ok {
		var err error
		rv := &sEither{make([]Selector, len(bcmp.Terms))}
		for i, b := range bcmp.Terms {
			rv.Terms[i], err = composeSelectors(top, b)
			if err != nil {
				return nil, err
			}
		}
		return rv, nil
	} else if btm, ok := bottom.(lefter); ok {
		btmleft := btm.Left()
		if btmleft.Type() == stImplicitAmp {
			if top == nil {
				// This is a top-level selector.
				// Remove implicit ampersand nodes from the selector
				bbt, ok := bottom.(*sCompound)
				if ok {
					return bbt.B.Clone(), nil
				} else {
					return bottom, compileError("Top-level selectors should be of the 'implicit descendant' type", nil)
				}
			}

			rv := bottom.Clone()
			if rrv, ok := rv.(*sCompound); ok {
				rrv.A = top.Clone()
			} else {
				return rv, compileError(fmt.Sprintf("Don't know how to compose type %d", rv.Type()), nil)
			}
			return rv, nil
		}
	}

	return top, compileError("Not implemented either", nil)
}
