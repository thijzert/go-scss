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
	fmt.Printf("%+v\n", err)
}
