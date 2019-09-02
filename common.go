package hammer

// Result holds the balance of each address
// All balances in Satoshi
type Result struct {
	Source, Address    string
	BalanceTotal       float64
	BalanceConfirmed   float64
	BalanceUnconfirmed float64
}

// W contains the fields that nearly all workers need
type W struct {
	name   string
	input  chan string
	output chan Result
	stop   chan struct{}
}

// Name of the worker
func (w W) Name() string {
	return w.name
}

// Worker defines the methods a balance querying worker must implement
type Worker interface {
	// Name of the worker
	Name() string
	// Start worker processing
	Start()
	// Stop worker processing
	Stop()
}

// Helpers

// WorkersAll returns slice containing all workers
func WorkersAll(input chan string, output chan Result, exit chan struct{}) []Worker {
	return []Worker{
		NewBlockonomics(input, output, exit),
		NewBlockcypher(input, output, exit),
	}
}

// SubmitAddresses submits addresses to a channel
// Usually called in a goroutine
func SubmitAddresses(addrs []string, ch chan string) {
	for _, addr := range addrs {
		ch <- addr
	}
}
