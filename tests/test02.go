// a test of terminal detection
package main

import (
	"fmt"
	"os"
)

func IsTTY() bool {
	if info, err := os.Stdout.Stat(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot stat stdout: %s\n", err.Error())
		return false
	} else if (info.Mode() & os.ModeDevice) != 0 {
		return true
	}
	return false
}

func main() {
	istty := IsTTY()
	fmt.Printf("istty = %v\n", istty)
	total := 10
	for i := 0; i < total; i++ {
		if istty {
			fmt.Printf("\r%d/%d", i, total)
		} else {
			fmt.Printf("%d/%d\n", i, total)
		}
		time.Sleep()
	}
	fmt.Println()
}
