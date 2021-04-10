# WaitGroupPool

[![GoDoc](https://godoc.org/github.com/remeh/waitgrouppool?status.svg)](https://godoc.org/github.com/remeh/waitgrouppool)

`WaitGroupPool` has the same role and API as `sync.WaitGroup` but it adds a limit of the amount of goroutines started concurrently.

`WaitGroupPool` adds the feature of limiting the maximum number of concurrently started routines. It could for example be used to start multiples routines querying a database but without sending too much queries in order to not overload the given database.

# Example

```go
package main

import (
        "fmt"
        "math/rand"
        "time"

        "github.com/quitz/waitgrouppool"
)

func main() {
        rand.Seed(time.Now().UnixNano())

        // Typical use-case:
        // 50 queries must be executed as quick as possible
        // but without overloading the database, so only
        // 8 routines should be started concurrently.
        swg := waitgrouppool.New(8)
        for i := 0; i < 50; i++ {
                swg.Add()
                go func(i int) {
                        defer swg.Done()
                        query(i)
                }(i)
        }

        swg.Wait()
}

func query(i int) {
        fmt.Println(i)
        ms := i + 500 + rand.Intn(500)
        time.Sleep(time.Duration(ms) * time.Millisecond)
}
```

# License

MIT

# Copyright

Rémy Mathieu © 2016
