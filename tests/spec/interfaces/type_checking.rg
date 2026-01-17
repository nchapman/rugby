#@ run-pass
#@ check-output
#
# Test: Section 13.4 - Type checking with is_a? and as

interface Runnable
  def run
end

class Worker implements Runnable
  def run
    puts "working"
  end
end

class Manager
  def manage
    puts "managing"
  end
end

def process(obj: Any)
  # is_a? returns Bool
  if obj.is_a?(Runnable)
    puts "is runnable"
  end

  # as returns T?
  if let r = obj.as(Runnable)
    r.run
  else
    puts "not runnable"
  end
end

process(Worker.new)
process(Manager.new)

#@ expect:
# is runnable
# working
# not runnable
