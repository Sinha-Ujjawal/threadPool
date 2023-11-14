package threadPool

import (
	"errors"
	"runtime"
)

type task[T any] struct {
	taskFn  func() T
	promise Promise[T]
}

type Promise[T any] struct {
	resultCh chan T
}

func (promise Promise[T]) Resolve() (*T, error) {
	result, ok := <-promise.resultCh
	if !ok {
		return nil, errors.New("Promise has been closed!")
	}
	return &result, nil
}

func ResolveMany[T any](promises []Promise[T]) ([]T, error) {
	ret := make([]T, len(promises))
	for i, promise := range promises {
		result, err := promise.Resolve()
		if err != nil {
			return nil, err
		}
		ret[i] = *result
	}
	return ret, nil
}

type ThreadPool[T any] struct {
	numWorkers uint32
	taskCh     chan task[T]
	isClosed   bool
}

func MkThreadPool[T any](numWorkers uint32) *ThreadPool[T] {
	if numWorkers == 0 {
		numWorkers = uint32(runtime.NumCPU())
	}
	tp := ThreadPool[T]{
		numWorkers: numWorkers,
		taskCh:     make(chan task[T]),
		isClosed:   false,
	}
	for _worker := uint32(0); _worker < numWorkers; _worker++ {
		go func() {
			for task := range tp.taskCh {
				task.promise.resultCh <- task.taskFn()
				close(task.promise.resultCh)
			}
		}()
	}
	return &tp
}

func (tp ThreadPool[T]) NumWorkers() uint32 {
	return tp.numWorkers
}

func (tp *ThreadPool[T]) Submit(taskFn func() T) (*Promise[T], error) {
	if tp.isClosed {
		return nil, errors.New("Threadpool is already closed!")
	}
	promise := Promise[T]{resultCh: make(chan T, 1)}
	go func() {
		task := task[T]{taskFn: taskFn, promise: promise}
		tp.taskCh <- task
	}()
	return &promise, nil
}

func Submit[T any, A any](tp *ThreadPool[T], mkTaskFn func(A) T, arg A) (*Promise[T], error) {
	return tp.Submit(func() T { return mkTaskFn(arg) })
}

func (tp *ThreadPool[T]) Close() {
	close(tp.taskCh)
	tp.isClosed = true
}
