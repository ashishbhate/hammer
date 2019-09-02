package hammer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const blockcypherURL = "https://api.blockcypher.com/v1/btc/main/addrs/%s/balance"
const blockcypherBatchLimit = 3    // can query upto this many addresses at time
const blockcypherHourlyLimit = 200 // can query upto this many addresses per hour

type blockcypherResponse struct {
	Addr        string  `json:"address"`
	Confirmed   float64 `json:"balance"`
	Unconfirmed float64 `json:"unconfirmed_balance"`
	Total       float64 `json:"final_balance"`
}

// Blockcypher worker
type Blockcypher struct {
	W
	currentHourlyCount int          // upto 200 queries per top of the UTC hour
	lastUTCReset       time.Time    // Last time counter was reset
	hourlyUTCTicker    *time.Ticker // Ticks every UTC hour
}

// NewBlockcypher returns an initialized Blockcypher worker
func NewBlockcypher(input chan string, output chan Result, stop chan struct{}) *Blockcypher {
	now := time.Now().UTC()
	bc := Blockcypher{
		W: W{
			name:   "blockcypher",
			input:  input,
			output: output,
			stop:   stop,
		},
		currentHourlyCount: 0,
		lastUTCReset:       now,
		hourlyUTCTicker:    nextUTCHourTicker(now),
	}
	return &bc
}

// Start the Blockcypher worker
func (bc *Blockcypher) Start() {
	addresses := make([]string, 0, blockcypherBatchLimit)
	for {
		if bc.currentHourlyCount == blockcypherHourlyLimit {
			now := time.Now().UTC()
			time.Sleep(durationTillNextUTCHour(now))
		}
		// we wait upto 5 seconds to gather as many addresses (upto batch limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case address := <-bc.input:
			addresses = append(addresses, address)
			if len(addresses) == blockcypherBatchLimit ||
				len(addresses) == blockcypherHourlyLimit-bc.currentHourlyCount {
				bc.currentHourlyCount += len(addresses)
				bc.process(addresses)
				addresses = []string{}
			}
		case <-ticker.C:
			if len(addresses) > 0 {
				bc.process(addresses)
				addresses = []string{}
				bc.currentHourlyCount += len(addresses)
			}
		case <-bc.hourlyUTCTicker.C:
			bc.hourlyUTCTicker.Stop()
			now := time.Now().UTC()
			bc.lastUTCReset = now
			bc.hourlyUTCTicker = nextUTCHourTicker(now)
			bc.currentHourlyCount = 0
		case <-bc.stop:
			ticker.Stop()
			bc.Stop()
			return
		}
		// Blockcypher claims a rate limit of 3 requests/second
		// Our request and processing cycle likely takes more than 1 second,
		// but I still see rate limiting sometimes.
		time.Sleep(1 * time.Second)
	}
}

// Stop the Blockcyper worker
func (bc *Blockcypher) Stop() {
	bc.hourlyUTCTicker.Stop()
	return
}

func (bc *Blockcypher) do(addresses []string) ([]blockcypherResponse, error) {
	addrs := strings.Join(addresses, ";")
	resp, err := http.Get(fmt.Sprintf(blockcypherURL, addrs))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 429 { // we've been rate limited for some reason
			bc.currentHourlyCount = blockcypherHourlyLimit
			return nil,
				fmt.Errorf("rate limited by blockcypher, got status code: %d", resp.StatusCode)
		}
		return nil,
			fmt.Errorf("error response from blockcypher, got status code: %d", resp.StatusCode)
	}

	var result []blockcypherResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (bc *Blockcypher) process(addresses []string) {
	resp, err := bc.do(addresses)
	if err != nil {
		fmt.Println(bc.Name()+":", err)
		go SubmitAddresses(addresses, bc.input) // return addresses to pool for processing
		return
	}
	for _, p := range resp {
		h := Result{
			Source:             bc.Name(),
			Address:            p.Addr,
			BalanceConfirmed:   p.Confirmed,
			BalanceUnconfirmed: p.Unconfirmed,
			BalanceTotal:       p.Total,
		}
		bc.output <- h
	}
}

func durationTillNextUTCHour(t time.Time) time.Duration {
	now := time.Now().UTC()
	nextUTCHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
	return nextUTCHour.Sub(now)
}

func nextUTCHourTicker(t time.Time) *time.Ticker {
	return time.NewTicker(durationTillNextUTCHour((t)))
}
