package main

import (
	"fmt"
	"os"

	"github.com/keidarcy/e1s/ui"
	"github.com/keidarcy/e1s/util"
)

func main() {
	if err := ui.Show(); err != nil {
		util.Logger.Printf("e1s - failed to start, error: %v\n", err)
		fmt.Println("e1s failed to start, please check your aws cli credential.")
		os.Exit(1)
	}
}
