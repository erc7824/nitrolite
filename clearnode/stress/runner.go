package stress

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	sdk "github.com/erc7824/nitrolite/sdk/go"
)

// RunTest executes totalReqs calls of fn distributed across the client pool.
func RunTest(ctx context.Context, totalReqs int, clients []*sdk.Client, fn MethodFunc) ([]Result, time.Duration) {
	numClients := len(clients)
	results := make([]Result, totalReqs)

	concurrency := totalReqs
	if numClients*10 < concurrency {
		concurrency = numClients * 10
	}
	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup
	var completed int64

	start := time.Now()

	for i := range totalReqs {
		sem <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			client := clients[idx%numClients]

			reqStart := time.Now()
			err := fn(ctx, client)
			d := time.Since(reqStart)
			results[idx] = Result{Duration: d, Err: err}

			c := atomic.AddInt64(&completed, 1)
			step := int64(totalReqs)/20 + 1
			if c%step == 0 || c == int64(totalReqs) {
				pct := float64(c) / float64(totalReqs) * 100
				fmt.Printf("\r  Progress: %d/%d (%.0f%%)  ", c, totalReqs, pct)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)
	fmt.Println()

	return results, totalTime
}
