#@ compile-fail
#
# Array.new without any arguments should error

def main
  arr = Array.new()  #~ ERROR: requires at least a size argument
end
