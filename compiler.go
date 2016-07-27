package scss

import (
	"fmt"
	"strings"
)

func Compile(src string) (string, error) {
	parseTree, err := Parse(src)
	if err != nil {
		errtext := err.Error()
		if perr, ok := err.(*ParseError); ok {
			errtext = perr.String()
		}

		return fmt.Sprintf("body:before { font-family: fixed; white-space: pre; content: \"%s\"; }", strings.Replace(strings.Replace(errtext, "\\", "\\\\", -1), "\"", "\\\"", -1)), compileError("Parse error", err)
	}

	rv := ""
	for _, r := range parseTree.Rules {
		rv += compileRule(r, "", "")
	}
	return rv, nil
}

func compileRule(rule Rule, previousRules, indent string) string {
	rv := ""

	previousRules += " " + rule.Selector.Evaluate()

	rv += indent + previousRules + " {\n"
	for _, p := range rule.Scope.Properties {
		rv += indent + "\t" + p.Key + ": " + p.Value + ";\n"
	}
	rv += indent + "}\n"

	for _, sr := range rule.Scope.Subrules {
		rv += compileRule(sr, previousRules, indent+"\t")
	}
	return rv
}
