package scss

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func Compile(src string) (string, error) {
	parseTree, err := Parse(src)
	if err != nil {
		errtext := err.Error()
		if perr, ok := err.(*ParseError); ok {
			errtext = perr.String()
		}

		return fmt.Sprintf("body:before { font-family: fixed; white-space: pre; content: \"%s\"; }", strings.Replace(strings.Replace(errtext, "\\", "\\\\", -1), "\"", "\\\"", -1)), errors.Wrap(err, "Parse error")
	}

	rv := ""
	for _, r := range parseTree.Rules {
		rv += compileRule(r, "", "") + "\n"
	}
	return rv, nil
}

func compileRule(rule Rule, previousRules, indent string) string {
	rv := ""

	previousRules += " "
	for _, s := range rule.Selector {
		if s.Type == WhitespaceToken {
			previousRules += " "
		} else {
			previousRules += s.Value
		}
	}

	rv += indent + previousRules + " {\n"
	for _, p := range rule.Scope.Properties {
		rv += indent + "   " + p.Key + ":" + p.Value + ";\n"
	}
	rv += indent + "}\n"

	for _, sr := range rule.Scope.Subrules {
		rv += compileRule(sr, previousRules, indent+"   ")
	}
	return rv
}
