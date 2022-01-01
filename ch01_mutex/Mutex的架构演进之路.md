**Mutex**

[TOC]



# 1. Mutex的架构演进之路

![image-20220101111939361](https://yfb-java.oss-cn-beijing.aliyuncs.com/%E7%AC%94%E8%AE%B0/image-20220101111939361.png)

## 1.1 初版

> 你可能会想到，可以通过一个 flag 变量，标记当前的锁是否被某个 goroutine 持有。如果
> 这个 flag 的值是 1，就代表锁已经被持有，那么，其它竞争的 goroutine 只能等待；如果
> 这个 flag 的值是 0，就可以通过 CAS（compare-and-swap，或者 compare-and-set）
> 将这个 flag 设置为 1，标识锁被当前的这个 goroutine 持有了。

这里涉及到cas方面的知识，不知道的话可以先去补充一下。

2008年提交的第一版mutex代码：

```go
// CAS操作，当时还没有抽象出atomic包
func cas(val *int32, old, new int32) bool
func semacquire(*int32)
func semrelease(*int32)
// 互斥锁的结构，包含两个字段
type Mutex struct {
	key int32 // 字段 key：是一个 flag，用来标识这个排外锁是否被某个 goroutine 所持有，如果 key大于等于 1，说明这个排外锁已经被持有；
	sema int32 // 字段 sema：是个信号量变量，用来控制等待 goroutine 的阻塞休眠和唤醒。
}
// 保证成功在val上增加delta的值
func xadd(val *int32, delta int32) (new int32) {
	for {
		v := *val
		if cas(val, v, v+delta) {
			return v + delta
		}
	}
	panic("unreached")
}
// 请求锁
func (m *Mutex) Lock() {
	if xadd(&m.key, 1) == 1 { //标识加1，如果等于1，成功获取到锁
		return
	}
	semacquire(&m.sema) // 否则阻塞等待
}
func (m *Mutex) Unlock() {
	if xadd(&m.key, -1) == 0 { // 将标识减去1，如果等于0，则没有其它等待者
		return
	}
	semrelease(&m.sema) // 唤醒其它阻塞的goroutine
}
```

![image-20220101114115315](C:\Users\51019\AppData\Roaming\Typora\typora-user-images\image-20220101114115315.png)

> Unlock 方法可以被任意的 goroutine 调用释放锁，即使是没持有这个互斥锁的
> goroutine，也可以进行这个操作。这是因为，Mutex 本身并没有包含持有这把锁的
> goroutine 的信息，所以，Unlock 也不会对此进行检查。Mutex 的这个设计一直保持至
> 今。

所以，这里我们为了避免在编写代码时错误的删除mutex中的Lock()或者UnLock()方法，一般我们会用defer这个方法来让LocK()和UnLock()方法成对出现，但是有时候临界区只是方法的一部分，我们为了及时的去释放锁，不会等到方法执行完才去释放锁，

```go
func(){
    sync.Mutex.Lock()
    defer sync.Mutex.UnLock()
}
```

## 1.2 给新人机会

> 但是，初版的 Mutex 实现有一个问题：请求锁的 goroutine 会排队等待获取互斥锁。虽
> 然这貌似很公平，但是从性能上来看，却不是最优的。因为如果我们能够把锁交给正在占
> 用 CPU 时间片的 goroutine 的话，那就不需要做上下文的切换，在高并发的情况下，可
> 能会有更好的性能。
>
> Go 开发者在 2011 年 6 月 30 日的 commit 中对 Mutex 做了一次大的调整，调整后的
> Mutex 实现如下：

```go
type Mutex struct {
	state int32
	sema uint32
}
const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexWaiterShift = iota
)
```

![image-20220101115531591](C:\Users\51019\AppData\Roaming\Typora\typora-user-images\image-20220101115531591.png)















---

# 2. Mutex使用中的4大易错场景

## 2.1 Lock/Unlock 不是成对出现
