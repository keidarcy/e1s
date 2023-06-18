package main

import (
	"fmt"

	"github.com/keidarcy/e1s/ui"
)

func main() {
	if err := ui.Show(); err != nil {
		fmt.Println("e1s failed to start, valid aws cli and aws cli profile are required")
		panic(err)
	}
}
