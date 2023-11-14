package main

import (
	"fmt"
	"threadPool"
	"time"
)

func main() {
	tp := threadPool.MkThreadPool[int](100)
	defer tp.Close()
	fmt.Printf("Num Workers: %d\n", tp.NumWorkers())
	now := time.Now()
	promises := []threadPool.Promise[int]{}
	for i := 0; i < 1000; i++ {
		taskFn := func(i int) int {
			time.Sleep(time.Millisecond * time.Duration(i+1))
			return i * i
		}
		promise, err := threadPool.Submit(tp, taskFn, i)
		if err != nil {
			panic(err)
		}
		promises = append(promises, *promise)
	}
	results, err := threadPool.ResolveMany(promises)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
	elapsed := time.Now().Sub(now)
	fmt.Printf("Time elapsed: %+v\n", elapsed)
}
