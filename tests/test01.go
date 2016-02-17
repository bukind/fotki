// A test of parallel producers and a single consumer
package main

import (
	"fmt"
	"sync"
	"time"
)

const Delay = 200 * time.Millisecond

func main() {
	list := make([]int, 100)
	for i := 0; i < 100; i++ {
		list[i] = i
	}
	fmt.Println(list)

	type ResErr struct {
		result int
		err    error
	}

	// done := make(chan int)
	feed := make(chan int)
	resc := make(chan ResErr)

	var wg sync.WaitGroup
	const routines = 4
	for i := 0; i < routines; i++ {
		wg.Add(1)
		// read from feed, check done for completion
		go func(idx int, // done <-chan int,
			feed <-chan int, resc chan<- ResErr) {
			fmt.Printf("func #%d STARTED\n", idx)
			defer wg.Done()
			defer fmt.Printf("func #%d EXITING\n", idx)
			for value := range feed {

				// leading wait
				select {
				// case <-done: return
				case <-time.After(Delay):
					// continue
				}

				tmp := ResErr{value * value, nil}
				if (value+idx)%routines == 0 {
					tmp.result = value
					tmp.err = fmt.Errorf("bad value %d", value)
				}

				select {
				// case <-done: return
				case resc <- tmp:
					// continue
				}
				fmt.Printf("func #%d done %d\n", idx, value)
			}
		}(i, // done,
			feed, resc)
	}

	// start consumer
	wg.Add(1)
	go func( // done <-chan int,
		resc <-chan ResErr) {
		defer wg.Done()
		defer fmt.Println("consumer EXITING")
		fmt.Println("consumer STARTED")
		for {
			select {
			// case <-done: return
			case res := <-resc:
				if res.err != nil {
					fmt.Printf("value #%d -> %s\n", res.result, res.err.Error())
				} else {
					fmt.Printf("result %d\n", res.result)
				}
			}
		}
	}( // done,
		resc)

	// pass data to the routines
	// tmo := time.After(2*time.Second)
	// feed:
	for _, value := range list {
		select {
		// case <-tmo: break feed
		case feed <- value:
			// continue
		}
	}
	close(feed)

	fmt.Println("closing done")
	// close(done)

	wg.Wait()
}
