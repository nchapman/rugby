#@ run-pass
#@ check-output
#
# Test: Section 13.1 - Interface definition

interface Reader
  def read(size : Int) -> String
end

interface Writer
  def write(data : String) -> Int
end

class StringBuffer implements Reader, Writer
  def initialize
    @buffer = ""
  end

  def read(size : Int) -> String
    result = @buffer[0...size]
    @buffer = @buffer[size..]
    result
  end

  def write(data : String) -> Int
    @buffer = @buffer + data
    data.length
  end
end

buf = StringBuffer.new
buf.write("hello world")
puts buf.read(5)
puts buf.read(6)

#@ expect:
# hello
#  world
