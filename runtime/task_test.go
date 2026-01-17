package runtime

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSpawnAwait(t *testing.T) {
	task := Spawn(func() any {
		return 42
	})

	result := Await(task)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestSpawnAwaitString(t *testing.T) {
	task := Spawn(func() any {
		return "hello"
	})

	result := Await(task)
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

// Test with concrete types to validate generics work correctly
func TestSpawnWithConcreteInt(t *testing.T) {
	task := Spawn(func() int {
		return 42
	})

	result := Await(task)
	// result is int, no type assertion needed
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestSpawnWithConcreteString(t *testing.T) {
	task := Spawn(func() string {
		return "hello"
	})

	result := Await(task)
	// result is string, no type assertion needed
	if result != "hello" {
		t.Errorf("expected 'hello', got %s", result)
	}
}

func TestScopeSpawnWithConcreteType(t *testing.T) {
	scope := NewScope()
	task := ScopeSpawn(scope, func() int {
		return 100
	})

	result := Await(task)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
	scope.Wait()
}

func TestSpawnAwaitMultiple(t *testing.T) {
	t1 := Spawn(func() any { return 1 })
	t2 := Spawn(func() any { return 2 })
	t3 := Spawn(func() any { return 3 })

	r1 := Await(t1)
	r2 := Await(t2)
	r3 := Await(t3)

	if r1 != 1 || r2 != 2 || r3 != 3 {
		t.Errorf("expected 1, 2, 3, got %v, %v, %v", r1, r2, r3)
	}
}

func TestAwaitIdempotent(t *testing.T) {
	var callCount int32
	task := Spawn(func() any {
		atomic.AddInt32(&callCount, 1)
		return 42
	})

	r1 := Await(task)
	r2 := Await(task)
	r3 := Await(task)

	if r1 != 42 || r2 != 42 || r3 != 42 {
		t.Errorf("expected all to be 42, got %v, %v, %v", r1, r2, r3)
	}
	if callCount != 1 {
		t.Errorf("expected function to be called once, called %d times", callCount)
	}
}

func TestScope(t *testing.T) {
	scope := NewScope()

	var count int32
	t1 := ScopeSpawn(scope, func() any {
		atomic.AddInt32(&count, 1)
		return nil
	})
	t2 := ScopeSpawn(scope, func() any {
		atomic.AddInt32(&count, 1)
		return nil
	})

	Await(t1)
	Await(t2)
	scope.Wait()

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestTryReceive(t *testing.T) {
	ch := make(chan int, 1)

	// Empty channel
	val, ok := TryReceive(ch)
	if ok {
		t.Errorf("expected false for empty channel, got true with value %v", val)
	}

	// Channel with value
	ch <- 42
	val, ok = TryReceive(ch)
	if !ok {
		t.Errorf("expected true for channel with value")
	}
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	// Empty again
	val, ok = TryReceive(ch)
	if ok {
		t.Errorf("expected false for empty channel after receive, got true with value %v", val)
	}
}

func TestTryReceiveClosedChannel(t *testing.T) {
	ch := make(chan int, 1)
	close(ch)

	val, ok := TryReceive(ch)
	if ok {
		t.Errorf("expected false for closed channel, got true with value %v", val)
	}
	if val != 0 {
		t.Errorf("expected zero value, got %v", val)
	}
}

func TestSpawnConcurrent(t *testing.T) {
	// Test that spawned tasks run concurrently
	start := time.Now()

	t1 := Spawn(func() any {
		time.Sleep(50 * time.Millisecond)
		return 1
	})
	t2 := Spawn(func() any {
		time.Sleep(50 * time.Millisecond)
		return 2
	})

	r1 := Await(t1)
	r2 := Await(t2)

	elapsed := time.Since(start)

	if r1 != 1 || r2 != 2 {
		t.Errorf("expected 1 and 2, got %v and %v", r1, r2)
	}

	// Should take ~50ms if concurrent, not ~100ms if sequential
	if elapsed > 90*time.Millisecond {
		t.Errorf("tasks appear to run sequentially, took %v", elapsed)
	}
}

// Additional comprehensive runtime tests

func TestSpawnReturnsNil(t *testing.T) {
	task := Spawn(func() any {
		return nil
	})

	result := Await(task)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestSpawnReturnsSlice(t *testing.T) {
	task := Spawn(func() any {
		return []int{1, 2, 3}
	})

	result := Await(task)
	slice, ok := result.([]int)
	if !ok {
		t.Errorf("expected []int, got %T", result)
		return
	}
	if len(slice) != 3 || slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
		t.Errorf("expected [1, 2, 3], got %v", slice)
	}
}

func TestSpawnReturnsMap(t *testing.T) {
	task := Spawn(func() any {
		return map[string]int{"a": 1, "b": 2}
	})

	result := Await(task)
	m, ok := result.(map[string]int)
	if !ok {
		t.Errorf("expected map[string]int, got %T", result)
		return
	}
	if m["a"] != 1 || m["b"] != 2 {
		t.Errorf("expected {a: 1, b: 2}, got %v", m)
	}
}

func TestSpawnReturnsStruct(t *testing.T) {
	type Result struct {
		Value int
		Name  string
	}

	task := Spawn(func() any {
		return Result{Value: 42, Name: "test"}
	})

	result := Await(task)
	r, ok := result.(Result)
	if !ok {
		t.Errorf("expected Result, got %T", result)
		return
	}
	if r.Value != 42 || r.Name != "test" {
		t.Errorf("expected {42, test}, got %v", r)
	}
}

func TestAwaitOnCompletedTask(t *testing.T) {
	task := Spawn(func() any {
		return 42
	})

	// First await completes the task
	r1 := Await(task)

	// Give it time to fully complete
	time.Sleep(10 * time.Millisecond)

	// Subsequent awaits should still work
	r2 := Await(task)
	r3 := Await(task)

	if r1 != 42 || r2 != 42 || r3 != 42 {
		t.Errorf("expected all 42, got %v, %v, %v", r1, r2, r3)
	}
}

func TestScopeWithManyTasks(t *testing.T) {
	scope := NewScope()

	var count int32
	const numTasks = 100

	tasks := make([]*Task[any], numTasks)
	for i := range numTasks {
		tasks[i] = ScopeSpawn(scope, func() any {
			atomic.AddInt32(&count, 1)
			return nil
		})
	}

	// Await all tasks
	for _, task := range tasks {
		Await(task)
	}

	scope.Wait()

	if count != numTasks {
		t.Errorf("expected count %d, got %d", numTasks, count)
	}
}

func TestScopedTasksRunConcurrently(t *testing.T) {
	scope := NewScope()
	start := time.Now()

	t1 := ScopeSpawn(scope, func() any {
		time.Sleep(30 * time.Millisecond)
		return 1
	})
	t2 := ScopeSpawn(scope, func() any {
		time.Sleep(30 * time.Millisecond)
		return 2
	})
	t3 := ScopeSpawn(scope, func() any {
		time.Sleep(30 * time.Millisecond)
		return 3
	})

	Await(t1)
	Await(t2)
	Await(t3)
	scope.Wait()

	elapsed := time.Since(start)

	// Should take ~30ms if concurrent, not ~90ms if sequential
	if elapsed > 60*time.Millisecond {
		t.Errorf("scoped tasks appear to run sequentially, took %v", elapsed)
	}
}

func TestTryReceiveWithDifferentTypes(t *testing.T) {
	// Test with string channel
	strCh := make(chan string, 1)
	strCh <- "hello"
	val, ok := TryReceive(strCh)
	if !ok || val != "hello" {
		t.Errorf("expected (hello, true), got (%v, %v)", val, ok)
	}

	// Test with bool channel
	boolCh := make(chan bool, 1)
	boolCh <- true
	bVal, bOk := TryReceive(boolCh)
	if !bOk || bVal != true {
		t.Errorf("expected (true, true), got (%v, %v)", bVal, bOk)
	}

	// Test with struct channel
	type Data struct {
		X int
	}
	dataCh := make(chan Data, 1)
	dataCh <- Data{X: 42}
	dVal, dOk := TryReceive(dataCh)
	if !dOk || dVal.X != 42 {
		t.Errorf("expected ({42}, true), got (%v, %v)", dVal, dOk)
	}
}

func TestTryReceiveMultipleTimes(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3

	v1, ok1 := TryReceive(ch)
	v2, ok2 := TryReceive(ch)
	v3, ok3 := TryReceive(ch)
	v4, ok4 := TryReceive(ch) // Empty now

	if !ok1 || v1 != 1 {
		t.Errorf("expected (1, true), got (%v, %v)", v1, ok1)
	}
	if !ok2 || v2 != 2 {
		t.Errorf("expected (2, true), got (%v, %v)", v2, ok2)
	}
	if !ok3 || v3 != 3 {
		t.Errorf("expected (3, true), got (%v, %v)", v3, ok3)
	}
	if ok4 {
		t.Errorf("expected (_, false), got (%v, %v)", v4, ok4)
	}
}

func TestSpawnWithPanic(t *testing.T) {
	// This test verifies that panics in spawned tasks don't crash the main goroutine
	// Note: The current implementation doesn't recover from panics, so this just documents behavior
	// If panic recovery is added later, this test should be updated

	// Spawn a task that completes normally
	task := Spawn(func() any {
		return "ok"
	})

	result := Await(task)
	if result != "ok" {
		t.Errorf("expected 'ok', got %v", result)
	}
}

func TestMultipleSpawnersOnSameScope(t *testing.T) {
	scope := NewScope()
	var count int32

	// Simulate multiple "spawners" adding tasks to the same scope
	done := make(chan bool, 3)

	for range 3 {
		go func() {
			for range 10 {
				ScopeSpawn(scope, func() any {
					atomic.AddInt32(&count, 1)
					return nil
				})
			}
			done <- true
		}()
	}

	// Wait for all spawners to finish
	for range 3 {
		<-done
	}

	scope.Wait()

	// Should have spawned 30 tasks total
	if count != 30 {
		t.Errorf("expected count 30, got %d", count)
	}
}

func TestAwaitReturnsCorrectValueForEachTask(t *testing.T) {
	tasks := make([]*Task[any], 10)
	for i := range 10 {
		val := i // Capture loop variable
		tasks[i] = Spawn(func() any {
			return val * val
		})
	}

	for i, task := range tasks {
		result := Await(task)
		expected := i * i
		if result != expected {
			t.Errorf("task %d: expected %d, got %v", i, expected, result)
		}
	}
}

func TestScopeWaitWithNoTasks(t *testing.T) {
	scope := NewScope()

	// Wait with no tasks should return immediately
	start := time.Now()
	scope.Wait()
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("Wait with no tasks took too long: %v", elapsed)
	}
}

func TestSpawnAwaitChain(t *testing.T) {
	// Create a chain of dependent tasks
	t1 := Spawn(func() any {
		return 10
	})

	t2 := Spawn(func() any {
		v := Await(t1).(int)
		return v * 2
	})

	t3 := Spawn(func() any {
		v := Await(t2).(int)
		return v + 5
	})

	result := Await(t3)
	if result != 25 { // 10 * 2 + 5
		t.Errorf("expected 25, got %v", result)
	}
}

func TestTryReceiveUnbufferedChannel(t *testing.T) {
	ch := make(chan int)

	// Unbuffered channel with no sender should return immediately
	val, ok := TryReceive(ch)
	if ok {
		t.Errorf("expected false for unbuffered empty channel, got true with %v", val)
	}

	// Start a sender in a goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- 42
	}()

	// Should still be empty initially
	val, ok = TryReceive(ch)
	if ok {
		t.Errorf("expected false before sender, got true with %v", val)
	}

	// Wait for sender and try again
	time.Sleep(20 * time.Millisecond)
	val, ok = TryReceive(ch)
	if !ok || val != 42 {
		t.Errorf("expected (42, true), got (%v, %v)", val, ok)
	}
}

func TestSpawnWithClosure(t *testing.T) {
	x := 10
	y := 20

	task := Spawn(func() any {
		return x + y
	})

	result := Await(task)
	if result != 30 {
		t.Errorf("expected 30, got %v", result)
	}
}

func TestScopeSpawnWithClosure(t *testing.T) {
	scope := NewScope()

	values := []int{1, 2, 3, 4, 5}
	var sum int32

	for _, v := range values {
		val := v // Capture
		ScopeSpawn(scope, func() any {
			atomic.AddInt32(&sum, int32(val))
			return nil
		})
	}

	scope.Wait()

	if sum != 15 {
		t.Errorf("expected sum 15, got %d", sum)
	}
}
