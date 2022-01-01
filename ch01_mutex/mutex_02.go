package main

import (
	"fmt"
	"sync"
)

// z针对mutex_01出现的问题加上mutex来锁住临界区
func main() {
	// 互斥锁用来保护计数器
	var mu sync.Mutex
	var count = 0
	// 辅助变量，用来确认所有的gourontine都完成
	var wg sync.WaitGroup
	wg.Add(10)
	// 启动10个gourontine
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Print(count)
}
