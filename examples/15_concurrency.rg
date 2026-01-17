# Rugby Concurrency
# Demonstrates: go, Chan, spawn/await, select, concurrently
#
# Rugby makes concurrency approachable. Choose the right tool:
# - go: fire-and-forget
# - spawn/await: get a result back
# - concurrently: parallel ops with guaranteed cleanup
# - Chan + select: low-level communication

import time
import sync

def main
  puts "=== Goroutines (fire-and-forget) ==="
  goroutine_demo

  puts "\n=== Channels ==="
  channel_demo

  puts "\n=== Spawn/Await (get result back) ==="
  spawn_demo

  puts "\n=== Select ==="
  select_demo

  puts "\n=== Structured Concurrency ==="
  concurrently_demo
end

# go - fire and forget background work
def goroutine_demo
  wg = sync.WaitGroup.new
  wg.add(3)

  go do
    time.sleep(10 * time.Millisecond)
    puts "  Worker 1 done"
    wg.done
  end

  go do
    time.sleep(5 * time.Millisecond)
    puts "  Worker 2 done"
    wg.done
  end

  go do
    time.sleep(15 * time.Millisecond)
    puts "  Worker 3 done"
    wg.done
  end

  wg.wait
  puts "  All workers finished"
end

# Channels for communication
def channel_demo
  ch = Chan<Int>.new(3)

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

# spawn/await - run work, get result
def spawn_demo
  task = spawn do
    time.sleep(10 * time.Millisecond)
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
  time.sleep(5 * time.Millisecond)
  n * 2
end

# select - wait for first ready channel
def select_demo
  ch1 = Chan<String>.new(1)
  ch2 = Chan<String>.new(1)

  go do
    time.sleep(10 * time.Millisecond)
    ch1 << "from ch1"
  end

  go do
    time.sleep(5 * time.Millisecond)
    ch2 << "from ch2"
  end

  select
  when msg = ch1.receive
    puts "  Got: #{msg}"
  when msg = ch2.receive
    puts "  Got: #{msg}"
  end
end

# concurrently - structured concurrency with cleanup
def concurrently_demo
  result = concurrently do |scope|
    t1 = scope.spawn -> { compute(10) }
    t2 = scope.spawn -> { compute(20) }
    t3 = scope.spawn -> { compute(30) }

    r1 = await t1
    r2 = await t2
    r3 = await t3

    r1 + r2 + r3
  end
  puts "  Concurrently result: #{result}"
end
