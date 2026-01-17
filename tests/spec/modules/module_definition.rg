#@ run-pass
#@ check-output
#
# Test: Section 14.1 - Module definition and include

module Loggable
  def log(msg: String)
    puts "[LOG] #{msg}"
  end

  def debug(msg: String)
    puts "[DEBUG] #{msg}"
  end
end

class Worker
  include Loggable

  def work
    log("Starting work")
    debug("Doing something")
    log("Work complete")
  end
end

worker = Worker.new
worker.work

#@ expect:
# [LOG] Starting work
# [DEBUG] Doing something
# [LOG] Work complete
