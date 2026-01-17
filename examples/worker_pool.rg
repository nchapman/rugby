# Worker Pool - Idiomatic concurrency pattern
# Demonstrates: channels, goroutines, structured iteration

import time

class Job
  getter id : Int
  getter payload : String

  def initialize(@id : Int, @payload : String)
  end
end

class Result
  getter job_id : Int
  getter output : String
  getter worker : Int

  def initialize(@job_id : Int, @output : String, @worker : Int)
  end
end

def process_job(job : Job, worker_id : Int) -> Result
  # Simulate work
  time.sleep(50 * time.Millisecond)
  output = job.payload.upcase
  Result.new(job.id, output, worker_id)
end

def worker(id : Int, jobs : Chan<Job>, results : Chan<Result>)
  for job in jobs
    result = process_job(job, id)
    results << result
  end
end

def main
  num_workers = 3
  num_jobs = 10

  jobs = Chan<Job>.new(num_jobs)
  results = Chan<Result>.new(num_jobs)

  # Start workers
  puts "Starting #{num_workers} workers..."
  num_workers.times -> do |i|
    go worker(i + 1, jobs, results)
  end

  # Send jobs
  puts "Sending #{num_jobs} jobs..."
  num_jobs.times -> do |i|
    job = Job.new(i + 1, "task-#{i + 1}")
    jobs << job
  end
  jobs.close

  # Collect results
  puts "\nResults:"
  num_jobs.times do
    result = results.receive
    puts "  Job ##{result.job_id}: #{result.output} (worker #{result.worker})"
  end

  puts "\nDone!"
end
