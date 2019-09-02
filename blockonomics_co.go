package hammer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const blockonomicsURL = "https://www.blockonomics.co/api/balance"
const blockonomicsBatchLimit = 25 // can query upto this many addresses at time

type blockonomicsBalances struct {
	Addr        string  `json:"addr"`
	Confirmed   float64 `json:"confirmed"`
	Unconfirmed float64 `json:"unconfirmed"`
}

type blockonomicsResponse struct {
	Response []blockonomicsBalances `json:"response"`
}

// Blockonomics worker
type Blockonomics struct {
	W
}

// NewBlockonomics returns an initialized Blockonomics worker
func NewBlockonomics(input chan string, output chan Result, stop chan struct{}) *Blockonomics {
	return &Blockonomics{
		W{
			name:   "blockonomics",
			input:  input,
			output: output,
			stop:   stop,
		},
	}
}

// Start the blockonomics worker
func (bl *Blockonomics) Start() {
	addresses := make([]string, 0, blockonomicsBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case address := <-bl.input:
			addresses = append(addresses, address)
			if len(addresses) == blockonomicsBatchLimit {
				bl.process(addresses)
				addresses = []string{}
			}
		case <-ticker.C:
			if len(addresses) > 0 {
				bl.process(addresses)
				addresses = []string{}
			}
		case <-bl.stop:
			ticker.Stop()
			bl.Stop()
			return
		}
	}
}

// Stop the Blockononics worker
func (bl *Blockonomics) Stop() {
	return
}

func (bl *Blockonomics) do(addresses []string) (blockonomicsResponse, error) {
	addrs := strings.Join(addresses, " ")
	req, err := json.Marshal(map[string]string{
		"addr": addrs,
	})
	if err != nil {
		return blockonomicsResponse{}, err
	}

	resp, err := http.Post(blockonomicsURL,
		"application/json",
		bytes.NewBuffer(req))
	if err != nil {
		return blockonomicsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return blockonomicsResponse{},
			fmt.Errorf("error response from blockonomics, got status code: %q", resp.StatusCode)
	}

	var result blockonomicsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (bl *Blockonomics) process(addresses []string) {
	resp, err := bl.do(addresses)
	if err != nil {
		fmt.Println(bl.Name()+":", err)
		go SubmitAddresses(addresses, bl.input) // return addresses to pool for processing
		return
	}
	for _, p := range resp.Response {
		h := Result{
			Source:             bl.Name(),
			Address:            p.Addr,
			BalanceConfirmed:   p.Confirmed,
			BalanceUnconfirmed: p.Unconfirmed,
			BalanceTotal:       p.Confirmed + p.Unconfirmed,
		}
		bl.output <- h
	}
}
