// Package runtime provides Ruby-like ergonomics for Go.
package runtime

import (
	"context"
	"sync"
)

// Task represents a concurrent computation that will produce a value.
// Tasks are created with Spawn and consumed with Await.
type Task struct {
	result chan any
	done   chan struct{}
	value  any
	once   sync.Once
}

// Spawn creates a new Task that executes the given function concurrently.
// The function is executed immediately in a goroutine.
func Spawn(fn func() any) *Task {
	t := &Task{
		result: make(chan any, 1),
		done:   make(chan struct{}),
	}

	go func() {
		defer close(t.done)
		t.result <- fn()
	}()

	return t
}

// Await blocks until the task completes and returns its value.
// Awaiting a completed task returns immediately.
func Await(t *Task) any {
	t.once.Do(func() {
		t.value = <-t.result
	})
	return t.value
}

// Scope represents a structured concurrency scope.
// All tasks spawned within a scope are awaited when the scope exits.
type Scope struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScope creates a new structured concurrency scope.
func NewScope() *Scope {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scope{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Spawn creates a new task within this scope.
// The task will be awaited when the scope exits.
func (s *Scope) Spawn(fn func() any) *Task {
	s.wg.Add(1)
	t := &Task{
		result: make(chan any, 1),
		done:   make(chan struct{}),
	}

	go func() {
		defer s.wg.Done()
		defer close(t.done)
		t.result <- fn()
	}()

	return t
}

// Ctx returns the scope's context for cancellation checking.
func (s *Scope) Ctx() context.Context {
	return s.ctx
}

// Wait waits for all spawned tasks to complete and cancels the context.
func (s *Scope) Wait() {
	s.cancel()
	s.wg.Wait()
}

// TryReceive attempts a non-blocking receive from a channel.
// Returns (value, true) if a value was received, (zero, false) otherwise.
func TryReceive[T any](ch <-chan T) (T, bool) {
	select {
	case v, ok := <-ch:
		if ok {
			return v, true
		}
		var zero T
		return zero, false
	default:
		var zero T
		return zero, false
	}
}

// Add adds two any values, handling int and float64 types.
// Returns the sum as any.
func Add(a, b any) any {
	switch av := a.(type) {
	case int:
		switch bv := b.(type) {
		case int:
			return av + bv
		case float64:
			return float64(av) + bv
		}
	case float64:
		switch bv := b.(type) {
		case int:
			return av + float64(bv)
		case float64:
			return av + bv
		}
	}
	panic("Add: unsupported types")
}
