// The executable program is the test and example
package main

import (
	"fmt"
	"sync"

	"gitlab.com/ashishbhate/hammer"
)

func main() {
	// Init
	output := make(chan hammer.Result)

	h := hammer.New(hammer.WorkersAll())

	// submit work
	// Although we know the number of addresses submitted
	// lets pretend we don't and use a wait group
	var wg sync.WaitGroup
	for _, addr := range hammer.SampleAddresses {
		wg.Add(1)
		go func(addr string) {
			output <- h.GetBalance(addr)
			wg.Done()
		}(addr)
	}

	// close the output at the right time
	go func() {
		wg.Wait()
		close(output)
	}()

	for res := range output {
		if res.BalanceTotal != 0 {
			fmt.Printf("%+v\n", res)
		}
	}
}
