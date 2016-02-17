// some path manipulation
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	path := "hello/world.ext"
	fmt.Println(path)
	fmt.Println(filepath.Base(path))
	fmt.Println(path[:len(path)-len(filepath.Ext(path))])
	noex := "bye/life"
	fmt.Println(noex)
	fmt.Println(noex[:len(noex)-len(filepath.Ext(noex))])
}
