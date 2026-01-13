# Rugby Concurrency
# Demonstrates: go, Chan, spawn/await, select, concurrently (spec 13)

import time
import sync

def main
  puts "=== Goroutines ==="
  goroutine_demo()

  puts "\n=== Channels ==="
  channel_demo()

  puts "\n=== Spawn/Await ==="
  spawn_demo()

  puts "\n=== Select ==="
  select_demo()

  puts "\n=== Structured Concurrency ==="
  concurrently_demo()
end

# Goroutines with go (spec 13.1)
def goroutine_demo
  wg = sync.WaitGroup.new
  wg.Add(3)

  go do
    time.Sleep(10 * time.Millisecond)
    puts "  Worker 1 done"
    wg.Done()
  end

  go do
    time.Sleep(5 * time.Millisecond)
    puts "  Worker 2 done"
    wg.Done()
  end

  go do
    time.Sleep(15 * time.Millisecond)
    puts "  Worker 3 done"
    wg.Done()
  end

  wg.Wait()
  puts "  All workers finished"
end

# Channels (spec 13.2)
def channel_demo
  ch = Chan[Int].new(3)

  # Send values
  ch << 10
  ch << 20
  ch << 30

  # Receive values
  v1 = ch.receive
  v2 = ch.receive
  v3 = ch.receive

  puts "  Received: #{v1}, #{v2}, #{v3}"
end

# spawn/await (spec 13.4)
def spawn_demo
  task = spawn do
    time.Sleep(10 * time.Millisecond)
    42
  end

  result = await task
  puts "  Task result: #{result}"

  # Multiple concurrent tasks
  t1 = spawn { compute(10) }
  t2 = spawn { compute(20) }
  t3 = spawn { compute(30) }

  r1 = await t1
  r2 = await t2
  r3 = await t3

  puts "  Results: #{r1}, #{r2}, #{r3}"
end

def compute(n : Int) -> Int
  time.Sleep(5 * time.Millisecond)
  n * 2
end

# select (spec 13.3)
def select_demo
  ch1 = Chan[String].new(1)
  ch2 = Chan[String].new(1)

  go do
    time.Sleep(10 * time.Millisecond)
    ch1 << "from ch1"
  end

  go do
    time.Sleep(5 * time.Millisecond)
    ch2 << "from ch2"
  end

  # Wait for first result
  select
  when msg = ch1.receive
    puts "  Got: #{msg}"
  when msg = ch2.receive
    puts "  Got: #{msg}"
  end
end

# Structured concurrency (spec 13.5)
# concurrently ensures all spawned tasks complete before exiting
def concurrently_demo
  result = concurrently do |scope|
    t1 = scope.spawn { compute(10) }
    t2 = scope.spawn { compute(20) }
    t3 = scope.spawn { compute(30) }

    r1 = await t1
    r2 = await t2
    r3 = await t3

    r1 + r2 + r3
  end
  puts "  Concurrently result: #{result}"
end
