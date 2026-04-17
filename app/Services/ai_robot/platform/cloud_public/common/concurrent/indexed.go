package concurrent

import "sync"

func RunIndexed(total int, maxConcurrent int, fn func(idx int)) {
	if total <= 0 || fn == nil {
		return
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn(idx)
		}(i)
	}
	wg.Wait()
}
