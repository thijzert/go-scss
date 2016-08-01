package scss

import (
	"fmt"
	"strings"
)

func Compile(src string) (string, error) {
	parseTree, err := Parse(src)
	if err != nil {
		return formatErrorCSS(err), compileError("Parse error", err)
	}

	rv := ""
	for _, r := range parseTree.Rules {
		rr, err := compileRule(r, nil, "")
		if err == nil {
			rv += rr
		} else {
			rv += formatErrorCSS(err)
			return rv, err
		}
	}
	return rv, nil
}

func formatErrorCSS(err error) string {
	errtext := err.Error()
	if perr, ok := err.(CompileError); ok {
		errtext = perr.String()
	} else if perr, ok := err.(ParseError); ok {
		errtext = perr.String()
	}
	return fmt.Sprintf("body:before { font-family: fixed; white-space: pre; content: \"%s\"; }", strings.Replace(strings.Replace(errtext, "\\", "\\\\", -1), "\"", "\\\"", -1))
}

func compileRule(rule Rule, prevSelector Selector, indent string) (string, error) {
	rv := ""

	thisSelector, err := composeSelectors(prevSelector, rule.Selector)
	if err != nil {
		return "", err
	}

	if len(rule.Scope.Properties) > 0 {
		rv += indent + thisSelector.Evaluate() + " {\n"
		for _, p := range rule.Scope.Properties {
			rv += indent + "\t" + p.Key + ": " + p.Value + ";\n"
		}
		rv += indent + "}\n"
	}

	for _, sr := range rule.Scope.Subrules {
		rr, err := compileRule(sr, thisSelector, indent+"\t")
		if err == nil {
			rv += rr
		} else {
			rv += formatErrorCSS(err)
			return rv, err
		}
	}

	return rv, nil
}
