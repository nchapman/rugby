#@ run-pass
#@ check-output
#
# Integration Test: State Machine Pattern
# Features: Enums + Case expressions + Classes + Methods + Implicit returns

# Order status enum
enum OrderStatus
  Pending = 1
  Confirmed = 2
  Shipped = 3
  Delivered = 4
  Cancelled = 5
end

# Order class implementing the state machine
class Order
  def initialize(@id: String, @status: OrderStatus)
  end

  getter id: String
  getter status: OrderStatus

  def can_advance?: Bool
    case @status
    when OrderStatus::Pending, OrderStatus::Confirmed, OrderStatus::Shipped
      true
    else
      false
    end
  end

  def advance: Bool
    return false unless can_advance?

    case @status
    when OrderStatus::Pending
      @status = OrderStatus::Confirmed
    when OrderStatus::Confirmed
      @status = OrderStatus::Shipped
    when OrderStatus::Shipped
      @status = OrderStatus::Delivered
    end
    true
  end

  def cancel: Bool
    if @status == OrderStatus::Pending || @status == OrderStatus::Confirmed
      @status = OrderStatus::Cancelled
      true
    else
      false
    end
  end

  def status_name: String
    @status.to_s
  end
end

# Helper to print order state
def print_order(order: Order)
  puts "Order #{order.id}: #{order.status_name}"
end

# Create orders
order1 = Order.new("ORD-001", OrderStatus::Pending)
order2 = Order.new("ORD-002", OrderStatus::Pending)

puts "Initial states:"
print_order(order1)
print_order(order2)

# Process order 1 through happy path
puts "Processing order 1:"
result = order1.advance
puts "  Advance to Confirmed: #{result}"
result = order1.advance
puts "  Advance to Shipped: #{result}"
result = order1.advance
puts "  Advance to Delivered: #{result}"
print_order(order1)

# Try to advance delivered order
puts "Try advancing delivered:"
result = order1.advance
puts "  Result: #{result}"

# Cancel order 2
puts "Cancelling order 2:"
result = order2.cancel
puts "  Cancel: #{result}"
print_order(order2)

# Try to cancel cancelled order
puts "Try cancelling again:"
result = order2.cancel
puts "  Result: #{result}"

# Test enum methods
puts "Enum features:"
puts "  Delivered value: #{OrderStatus::Delivered.value}"
puts "  Delivered to_s: #{OrderStatus::Delivered.to_s}"

#@ expect:
# Initial states:
# Order ORD-001: Pending
# Order ORD-002: Pending
# Processing order 1:
#   Advance to Confirmed: true
#   Advance to Shipped: true
#   Advance to Delivered: true
# Order ORD-001: Delivered
# Try advancing delivered:
#   Result: false
# Cancelling order 2:
#   Cancel: true
# Order ORD-002: Cancelled
# Try cancelling again:
#   Result: false
# Enum features:
#   Delivered value: 4
#   Delivered to_s: Delivered
