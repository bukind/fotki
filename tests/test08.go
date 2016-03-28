package main

import (
    "fmt"
	"os"
	"strconv"
	"unicode/utf8"
)


func main() {
    for _, arg := range os.Args[1:] {
	    start, err := strconv.ParseInt(arg, 0, 32)
		if err != nil {
		    fmt.Println("invalid argument %s\n", arg)
		}
		const rowsize = 32
		const nrows = 16
		start_row := int(start / rowsize)
		buf := make([]byte, 20)
		for row := start_row; row < start_row + nrows; row++ {
		    offset := row * rowsize
		    fmt.Printf("\n%6d/%4x", offset, offset)
		    for i := 0; i < rowsize; i++ {
			    r := rune(offset + i)
				cnt := utf8.EncodeRune(buf, r)
				fmt.Printf(" %s", string(buf[:cnt]))
			}
		}
	}
	fmt.Println()
}
