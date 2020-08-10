package main

import (
	"fmt"
	"os"
)

func main() {
	app := CLI()
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
