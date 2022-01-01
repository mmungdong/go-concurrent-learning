package main

import (
	"fmt"
	"sync"
)

// 线程安全的计数器类型
type Counterplus struct {
	CountType uint64
	Name      string

	mu    sync.Mutex
	count uint64
}

func (c *Counterplus) Incr() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *Counterplus) Count() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func main() {
	// 封装好的计数器
	var counter Counterplus
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				counter.Incr() // 受到保护的计数器方法
			}
		}()
	}
	wg.Wait()
	fmt.Print(counter.Count())
}
