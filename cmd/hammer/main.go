// The executable program is the test and example
package main

import (
	"fmt"

	"gitlab.com/ashishbhate/hammer"
)

func main() {
	input := make(chan string)
	output := make(chan hammer.Result)
	stop := make(chan struct{})

	workers := hammer.WorkersAll(input, output, stop)

	for _, worker := range workers {
		go worker.Start()
	}

	go hammer.SubmitAddresses(hammer.SampleAddresses, input)

	for range hammer.SampleAddresses {
		res := <-output
		fmt.Printf("%+v\n", res)
	}
	// signal exit by closing exit channel
}
