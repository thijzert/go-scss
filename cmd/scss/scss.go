package main

import (
	"fmt"
	"github.com/thijzert/go-scss"
)

func main() {
	src := `.foo
	{
		color: rgb(0, 100, 0);
		background-color: #000;
		.bar
		{
			color: lime;
		}
	}`

	parsed, err := scss.Parse(src)
	fmt.Printf("%+v\n", parsed)

	if err != nil {
		if perr, ok := err.(scss.ParseError); ok {
			fmt.Printf("%s\n", perr.String())
		} else {
			fmt.Printf("%+v\n", err)
		}
	}
	compiled, err := scss.Compile(src)
	fmt.Printf("%s\n", compiled)

	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
