# Goroutines
全ての Go プログラムは少なくとも1つの goroutine (= *main goroutine*)を持っている

ただ `go` と入れるだけでいい
```go
package main

import "fmt"

func main() {
	go sayHello()
	// continue doing other things
}

func sayHello() {
	fmt.Println("hello")
}
```
無名関数でもいい
```go
package main

import "fmt"

func main() {
	go func() {
		fmt.Println("hello")
	}() // すぐ実行しなければならない
}
```
その代わりに関数を変数に割り当てて後から呼んでもいい
```go
package main

import "fmt"

func main() {
	sayHello := func() {
		fmt.Println("hello")
	}
	go sayHello()
}
```

---

goroutine は OS のスレッドでもグリーンスレッドでもない  

||coroutine|goroutine|
|:---:|:---:|:---:|
|Multitasking|Nonpreemptive|Preemptive|
|Concurrency|※|✓|

---

goroutine の実装は M:N スケジューラと呼ばれる(6章で後述)

---

Go は *fork-join* モデルを採用している
![fork-join model](https://upload.wikimedia.org/wikipedia/commons/f/f1/Fork_join.svg)
(Source: https://en.wikipedia.org/wiki/Fork%E2%80%93join_model)

---

さっきの例に戻ってみると、join point が無いので恐らく何も出力されない
```go
package main

import "fmt"

func main() {
	sayHello := func() {
		fmt.Println("hello")
	}
	go sayHello()
}
```
goroutine は `sayHello` を実行するが、「ある」時間が経った後に終了する  
→ main goroutine が実行中に実行されないかもしれない
> goroutine 生成後に `time.Sleep` を実行しても、*join point* を作成することにはならないし、*race condition* の排除はできない

---

join point を作成するためには、main goroutine と `sayHello` goroutine をシンクロさせなければならない

方法はいくつかあるが、`sync.WaitGroup` を用いる
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("hello")
	}
	wg.Add(1)
	go sayHello()
	wg.Wait() // join point
}
```
実行結果
```sh
~/GoProject/concurrency_in_go/3/goroutines $  go run 3.go
hello
```
この例では、`sayHello` が終わるまで main goroutine がブロックされている

---

closure がスコープ外の変数にアクセスしている
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	salutation := "hello"
	wg.Add(1)
	go func() {
		defer wg.Done()
		salutation = "welcome"
	}()
	wg.Wait()
	fmt.Println(salutation)
}
```
実行結果
```sh
~/GoProject/concurrency_in_go/3/goroutines $  go run 4.go
welcome
```
この goroutine が main goroutine と同じアドレス空間を使用していることがわかる

---

この実行結果はどうなるだろうか
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation)
		}()
	}
	wg.Wait()
}
```
実行結果
```sh
~/GoProject/concurrency_in_go/3/goroutines $  go run 5_0.go
good day
good day
good day
```
goroutine はいつか実行されるものとしてスケジューリングされるので、for ループが終わるのが goroutine 実行より早かった可能性が高い  

goroutine が for ループが終わった後も `salutation` にアクセスできるとはどういうことか?  
Go ランタイムは `salutation` のリファレンスが保持されるとわかるのでヒープに移す  
→ goroutine からアクセスできる


ちなみに
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation)
		}()
		wg.Wait()
	}
}
```
実行結果
```sh
~/GoProject/concurrency_in_go/3/goroutines $  go run 5_1.go
hello
greetings
good day
```

---

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func(salutation string) {
			defer wg.Done()
			fmt.Println(salutation)
		}(salutation) // salutation のコピーが用意されたら実行される
	}
	wg.Wait()
}
```
実行結果
```sh
~/GoProject/concurrency_in_go/3/goroutines $  go run 6.go
good day
hello
greetings
~/GoProject/concurrency_in_go/3/goroutines $  go run 6.go
good day
greetings
hello
```
goroutine は同一のアドレス空間を用いて実行されるが、Go コンパイラはメモリ内の変数をうまく扱うので、goroutine が空きメモリにアクセスすることはない  
とはいえ依然 synchronization には気を使わなければならない(後述)

---

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	memConsumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys
	}

	var c <-chan interface{}
	var wg sync.WaitGroup
	noop := func() { wg.Done(); <-c }

	const numGoroutines = 1e4
	wg.Add(numGoroutines)
	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}
	wg.Wait()
	after := memConsumed()
	fmt.Printf("%.3fkb", float64(after-before)/numGoroutines/1000)
}
```
