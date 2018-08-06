# WaitGroup
次の場合におすすめ
* Concurrent なオペレーションの結果を気にしない
* 結果を集める別の手段がある(?)

そうでない場合、channel と `select` を使うのがオススメ

```go
package main

import (
	"sync"
	"fmt"
	"time"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2st goroutine sleeping...")
		time.Sleep(2)
	}()

	wg.Wait()
	fmt.Println("All goroutine complete.")
}
```
実行結果
```sh
2st goroutine sleeping...
1st goroutine sleeping...
All goroutine complete.
```
`WaitGroup` は concurrent-safe カウンタと見ることができる。  
`Add` でカウンタを増やし、`Done` でカウンタを減らし、`Wait` はカウンタが0になるまで待っている。

`Add` は goroutine の外で行われていることに留意しなければならない。仮にそうでない場合、race condition を考慮する羽目になる。  
goroutine はいつ実行されるかわからないため、`Add` が実行される前に `Wait` が実行される可能性が出る。

`Add` を goroutine の近くに置くのは慣習となっているが、こうすることもある
```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	hello := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		fmt.Printf("Hello from %v!\n", id)
	}

	const numGreeters = 5
	var wg sync.WaitGroup
	wg.Add(numGreeters)
	for i := 0; i < numGreeters; i++ {
		go hello(&wg, i + 1)
	}
	wg.Wait()
}
```
実行結果
```sh
Hello from 5!
Hello from 2!
Hello from 3!
Hello from 4!
Hello from 1!
```

# Mutex
"**mut**ual **ex**clusion"

`Mutex` は共有リソースの中で独占的な利用を提供できる。  
channel はメモリを通信によって共有するのに対して、`Mutex` はメモリを開発者が守るべき慣習によって共有している。

```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	var count int
	var lock sync.Mutex

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incrementing: %d\n", count)
	}

	decrement := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decrementing:: %d\n", count)

	}

	var arithmetic sync.WaitGroup
	for i := 0; i <= 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			increment()
		}()
	}

	for i := 0; i <= 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decrement()
		}()
	}

	arithmetic.Wait()
	fmt.Println("Arithmetic complete.")
}
```
実行結果
```
Decrementing: -1
Incrementing: 0
Incrementing: 1
Incrementing: 2
Incrementing: 3
Incrementing: 4
Incrementing: 5
Decrementing: 4
Decrementing: 3
Decrementing: 2
Decrementing: 1
Decrementing: 0
Arithmetic complete.
```

# RWMutex
`sync.RWMutex` は基本的には `Mutex` と同じだが、`RWMutex` はよりメモリに対してコントロールできることが増えている。

```go
package main

import (
	"sync"
	"time"
	"text/tabwriter"
	"os"
	"fmt"
	"math"
)

func main() {
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 5; i > 0; i-- {
			l.Lock()
			l.Unlock()
			time.Sleep(1)
		}
	}

	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	test := func(count int, mutex, rwMutex sync.Locker) time.Duration {
		var wg sync.WaitGroup
		wg.Add(count + 1)
		beginTestTime := time.Now()
		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rwMutex)
		}

		wg.Wait()
		return time.Since(beginTestTime)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := 0; i< 20; i++ {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(tw, "%d\t%v\t%v\n", count, test(count, &m, m.RLocker()), test(count, &m, &m),)
	}
}
```
実行結果
```sh
Readers  RWMutex       Mutex
1        24.009µs      4.012µs
2        4.758µs       3.161µs
4        72.717µs      12.909µs
8        14.742µs      32.972µs
16       35.705µs      15.981µs
32       65.078µs      11.391µs
64       95.002µs      75.267µs
128      149.945µs     58.231µs
256      126.435µs     86.059µs
512      88.511µs      119.934µs
1024     174.741µs     216.423µs
2048     339.647µs     363.185µs
4096     1.055861ms    1.21003ms
8192     1.633604ms    1.605446ms
16384    3.609927ms    5.026674ms
32768    6.246535ms    7.197636ms
65536    19.916885ms   46.80006ms
131072   38.27555ms    45.762824ms
262144   47.530145ms   153.997923ms
524288   193.936082ms  133.264763ms
```

# Cond
> ...a rendezvous point for goroutines waiting for or announcing the occurrence of an event.

"event" は複数の goroutine の間でやりとりされるシグナルで、それが発生したかどうか以上の情報を持たないものである。

ひとつのシグナルが来るのを待っていて、それを `Cond` なしで実現しようとするならば
```go
for conditionTrue() == false {
}
```
しかしこれは1つのコアの全計算資源を使ってしまう。
```go
for conditionTrue() == false {
	time.Sleep(1*time.Millisecond)
}
```
こうすればある程度は改善されるが、`time.Sleep` の時間を「ちょうどいい」時間にしなければならない。

`Cond` を使えば、次のように表現できる。
```go
c := sync.NewCond(&sync.Mutex{})
c.L.Lock()
for conditionTrue() == false {
	c.Wait()
}
c.L.Unlock()
```
さっきよりはかなり効率的である。  
`Wait` はただブロックしているのではなく、現在の goroutine をサスペンドさせているだけであるので、
他の goroutine は動作し続けることができる。

```go
package main

import (
	"sync"
	"time"
	"fmt"
)

func main() {
	c:= sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < 10; i++ {
		c.L.Lock()
		for len(queue) == 2 {
			c.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
	
}
```
実行結果
```sh
Adding to queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
Removed from queue
Adding to queue
```

## Signal, Broadcast
`Sygnal` は　`Wait` でブロックされている goroutine に通知するための機能。Go のランタイムは FIFO でシグナルを受ける goroutine を管理している。  
`Broadcast` は `Wait` でブロックされている goroutine 全てに対して通知を送る。

これらは channel で再現できるが、`Broadcast` を再現しようとする場合には `Cond` を使ったほうが圧倒的に効率がいい。

`Broadcast` は、あるボタンのクリックで複数の関数が走るような GUI アプリケーションを考えると便利に使える。

```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	type Button struct {
		Clicked *sync.Cond
	}
	button := Button{ Clicked: sync.NewCond(&sync.Mutex{})}

	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunning sync.WaitGroup
		goroutineRunning.Add(1)
		go func() {
			goroutineRunning.Done()
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait()
			fn()
		}()
		goroutineRunning.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing window.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked.")
		clickRegistered.Done()
	})

	button.Clicked.Broadcast()

	clickRegistered.Wait()
}
```
実行結果
```sh
Mouse clicked.
Maximizing window.
Displaying annoying dialog box!
```

# Once
次のコードは何を出力するだろうか。
```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	var count int

	increment := func() {
		count++
	}

	var once sync.Once

	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count)
}
```
実行結果
```sh
Count is 1
```

次はどうだろう。
```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	var count int

	increment := func() {
		count++
	}

	decrement := func() {
		count--
	}

	var once sync.Once

	var increments sync.WaitGroup
	var decrements sync.WaitGroup
	increments.Add(100)
	decrements.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
			once.Do(decrement)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count)
}
```
実行結果
```sh
Count is 1
```

次の例ではデッドロックを起こす。
```go
var onceA, onceB sync.Once
var initB func()
initA := func() { onceB.Do(initB) }
initB = func() { onceA.Do(initA) }
onceA.Do(initA)
```

# Pool
`Pool` は concurrent-safe なオブジェクトプールパターンである。
```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new instance.")
			return struct{}{}
		},
	}

	myPool.Get()
	instance := myPool.Get()
	myPool.Put(instance)
	myPool.Get()
}
```
実行結果
```
Creating new instance.
Creating new instance.
```

```go
package main

import (
	"sync"
	"fmt"
)

func main() {
	var numCalcsCreated int
	calcPool := &sync.Pool{
		New: func() interface{} {
			numCalcsCreated += 1
			mem := make([]byte, 1024)
			return &mem
		},
	}

	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	const numWorkers = 1024*1024
	var wg sync.WaitGroup
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem)
		}()
	}

	wg.Wait()
	fmt.Printf("%d calculators were created.", numCalcsCreated)
}
```
実行結果
```sh
panic: sync: negative WaitGroup counter

goroutine 44 [running]:
sync.(*WaitGroup).Add(0xc420084010, 0xffffffffffffffff)
	/usr/local/go/src/sync/waitgroup.go:73 +0x133
sync.(*WaitGroup).Done(0xc420084010)
	/usr/local/go/src/sync/waitgroup.go:98 +0x34
main.main.func2(0xc420084010, 0xc420098020)
	/Users/masahiko.hara/GoProject/ecfd_golang_study_meeting/concurrency_in_go/3/sync_package/8.go:31 +0x9c
created by main.main
	/Users/masahiko.hara/GoProject/ecfd_golang_study_meeting/concurrency_in_go/3/sync_package/8.go:26 +0x1a6
```
