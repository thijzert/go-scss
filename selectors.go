package scss

import (
	"fmt"
)

type selectorNodeType int

const (
	stNil selectorNodeType = iota
	stImplicitAmp
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

func (t selectorNodeType) String() string {
	if t == stNil {
		return "<nil>"
	} else if t == stImplicitAmp {
		return "ImplicitAmp"
	} else if t == stExplicitAmp {
		return "ExplicitAmp"
	} else if t == stCompoundBoth {
		return "CompoundBoth"
	} else if t == stCompoundEither {
		return "CompoundEither"
	} else if t == stCompoundDescendant {
		return "CompoundDescendant"
	} else if t == stCompoundDirectDescendant {
		return "CompoundDirectDescendant"
	} else if t == stCompoundNextSibling {
		return "CompoundNextSibling"
	} else if t == stID {
		return "ID"
	} else if t == stTag {
		return "Tag"
	} else if t == stClass {
		return "Class"
	} else if t == stPseudoclass {
		return "Pseudoclass"
	} else if t == stFunctionClass {
		return "FunctionClass"
	} else if t == stAttribute {
		return "Attribute"
	} else {
		return fmt.Sprintf("Unknown type %d")
	}
}

type Selector interface {
	Type() selectorNodeType
	Evaluate() string
	Clone() Selector
}

type lefter interface {
	Left() Selector
}

func parseSelector(tok *TokenRing) (rv Selector, err error) {
	rv, _, err = realParseSelector(tok, &sAmpersand{false})
	return
}

func realParseSelector(tok *TokenRing, left Selector) (rv Selector, explicitAmp bool, err error) {
	tok.Mark()
	peek := tok.Peek()
	whitespaceFound := peek != nil && peek.Type == WhitespaceToken
	explicitAmp = false

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
			return left, explicitAmp, nil
		}
	}
	if right.Type() == stExplicitAmp {
		explicitAmp = true
	}

	var farright Selector
	var amp bool
	farright, amp, err = realParseSelector(tok, right)
	explicitAmp = amp || explicitAmp
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
		} else if left.Type() == stImplicitAmp && explicitAmp {
			// No need for an implied amp node; we have one right here.
			rv = right
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
		} else if peek.Value == ":" {
			// Either a function class or a CSS2.1 pseudoclass
			peek = tok.Next()
			if peek == nil || peek.Type != SymbolToken {
				err = parseError("Expected symbol", nil, peek)
			} else {
				pp := tok.Peek()
				if pp != nil && pp.Type == SymbolToken && pp.Value == "(" {
					// TODO: "function class"
					err = parseError("Function classes (e.g. \":"+peek.Value+"(...)\" are not implemented", nil, peek)
				}
				rv = &sPseudoclass{peek.Value}
			}
		} else if peek.Value == "[" {
			// Attribute selector
			peek = tok.Next()
			if peek == nil || peek.Type != SymbolToken {
				err = parseError("[ Expected symbol", nil, peek)
			} else {
				rrv := &sAttribute{peek.Value, "", ""}
				peek = tok.Ignore(WhitespaceToken)
				if peek.Type != OperatorToken {
					err = parseError("[ Expected operator", nil, peek)
				} else {
					for peek.Type == OperatorToken {
						rrv.Operator += peek.Value
						peek = tok.Next()
					}
					tok.Rewind()
					peek = tok.Ignore(WhitespaceToken)
					if peek == nil || (peek.Type != SymbolToken && peek.Type != StringToken) {
						err = parseError("[ Expected: string or symbol; got '"+peek.Value+"'", nil, peek)
					} else {
						rrv.Value = peek.Value
						peek = tok.Ignore(WhitespaceToken)
						if peek == nil || peek.Type != OperatorToken || peek.Value != "]" {
							err = parseError("[ Expected: ']'", nil, peek)
						} else {
							rv = rrv
						}
					}
				}
			}
		} else if peek.Value == "&" {
			rv = &sAmpersand{true}
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

type sAmpersand struct {
	Explicit bool
}

func (s *sAmpersand) Type() selectorNodeType {
	if s.Explicit {
		return stExplicitAmp
	}
	return stImplicitAmp
}
func (s *sAmpersand) Evaluate() string {
	if s.Explicit {
		return "?!?"
	}
	return "?"
}
func (s *sAmpersand) Clone() Selector {
	return &sAmpersand{s.Explicit}
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

type sPseudoclass struct {
	Pseudoclass string
}

func (s *sPseudoclass) Type() selectorNodeType {
	return stPseudoclass
}
func (s *sPseudoclass) Evaluate() string {
	return ":" + s.Pseudoclass
}
func (s *sPseudoclass) Clone() Selector {
	return &sPseudoclass{s.Pseudoclass}
}

type sAttribute struct {
	AttributeName string
	Operator      string
	Value         string
}

func (s *sAttribute) Type() selectorNodeType {
	return stAttribute
}
func (s *sAttribute) Evaluate() string {
	return "[" + s.AttributeName + s.Operator + s.Value + "]"
}
func (s *sAttribute) Clone() Selector {
	return &sAttribute{s.AttributeName, s.Operator, s.Value}
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
	if bottom.Type() == stExplicitAmp {
		if top == nil {
			return bottom, compileError("Empty selector: composing <nil> and &", nil)
		}
		return top.Clone(), nil
	} else if top != nil && top.Type() == stExplicitAmp {
		if bottom == nil {
			return top, compileError("Empty selector: composing <nil> and &", nil)
		}
		return bottom.Clone(), nil
	} else if tcmp, ok := top.(*sEither); ok {
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
	}

	return applyAmpersand(top, bottom)
}

func applyAmpersand(amp, into Selector) (Selector, error) {
	icmp, icmpOK := into.(*sCompound)

	if amp == nil {
		// This is a top-level selector.
		// Remove implicit ampersand nodes from the selector

		if icmpOK && into.Type() == stCompoundDescendant {
			return icmp.B.Clone(), nil
		} else {
			return nil, compileError("Top-level selectors should be of the 'implicit descendant' type", nil)
		}
	}

	itype := into.Type()
	if itype == stImplicitAmp || itype == stExplicitAmp {
		return amp.Clone(), nil
	}
	if icmpOK {
		ca, ea := applyAmpersand(amp, icmp.A)
		cb, eb := applyAmpersand(amp, icmp.B)

		if ea != nil {
			return nil, compileError(fmt.Sprintf("Can't apply ampersand to type \"%s\" somehow", icmp.A.Type()), ea)
		}
		if eb != nil {
			return nil, compileError(fmt.Sprintf("Can't apply ampersand to type \"%s\" somehow", icmp.B.Type()), eb)
		}

		return &sCompound{icmp.Type(), ca, cb}, nil
	}

	return into.Clone(), nil
}
