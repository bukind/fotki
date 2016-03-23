// A test of utf8 identifiers

package main

import "fmt"

type привет string

func дай_пять() привет {
    return привет("привет")
}

func main() {
    fmt.Println(дай_пять())
}
