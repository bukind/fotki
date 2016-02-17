// test memory usage by strings
package main

import (
	"fmt"
	"runtime"
)

func useArray(array map[string]string) int {
	return len(array)
}

func showMem(oldm *runtime.MemStats, n uint64) *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory: HeapObjects=%d HeapInuse=%d StackInuse=%d\n",
		m.HeapObjects, m.HeapInuse, m.StackInuse)
	if oldm != nil {
		fmt.Printf("Delta:  HeapObjects=%d HeapInuse=%d StackInuse=%d\n",
			m.HeapObjects-oldm.HeapObjects,
			m.HeapInuse-oldm.HeapInuse,
			m.StackInuse-oldm.StackInuse)
		if n > 0 {
			fmt.Printf("PerObj: HeapObjects=%d HeapInuse=%d StackInuse=%d\n",
				(m.HeapObjects-oldm.HeapObjects)/n,
				(m.HeapInuse-oldm.HeapInuse)/n,
				(m.StackInuse-oldm.StackInuse)/n)
		}
	}
	return &m
}

func main() {
	m := showMem(nil, 0)
	var total uint64 = 500000
	array := make(map[string]string)
	var i uint64
	for i = 0; i < total; i++ {
		// the string must be 64 bytes + overhead
		s := fmt.Sprintf("%08d%08d%08d%08d%08d%08d%08d%08d%08d%08d%08d%08d%08d%08d",
			i, i, i, i, i, i, i, i, i, i, i, i, i, i)
		array[s] = s
	}
	showMem(m, total)
	for key, _ := range array {
		fmt.Printf("One string size=%d\n", len(key))
		break
	}
	_ = useArray(array)
}
