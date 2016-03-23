// A test of md5sum
package main

import (
    "fmt"
	"github.com/bukind/fotki"
)


func main() {
    md5, err := fotki.Md5sum("tests/test07.go")
	if err != nil {
	    fmt.Println("cannot md5:", err.Error())
	} else {
	    fmt.Printf("%+x\n",md5)
	}
}
