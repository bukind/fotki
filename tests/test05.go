// a test of fmt.Sscanf
package main

import (
	"fmt"
)

func check(line string) {
	if len(line) < 10 {
		fmt.Printf("Too short line: %d\n", len(line))
		return
	}
	var y, m, d int
	n, err := fmt.Sscanf(line[:10], "%04d-%02d-%02d", &y, &m, &d)
	if err != nil {
		fmt.Printf("Failed to scan %d: %s\n", n, err.Error())
		return
	}
	tail := line[10:]
	if len(tail) > 0 && tail[0] != '-' {
		fmt.Printf("Wrong tail for: %d %d %d %s\n", y, m, d, tail)
		return
	}
	fmt.Printf("Scanned %d items: %d %d %d %s\n", n, y, m, d, tail)
}

func main() {
	check("")
	check("2001")
	check("2001-01-01")
	check("2001-01-01-hello")
	check("2001-01-01badline")
	check("2001_01_01-hello")
	check("2001_01_01badline")
}
