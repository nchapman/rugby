#@ run-pass
#@ check-output
#
# Test: Section 14.2 - Module state (instance variables)

module Timestamped
  @created_at: Int?
  @updated_at: Int?

  def touch(now: Int)
    @created_at ||= now
    @updated_at = now
  end

  def created: Int
    @created_at ?? 0
  end

  def updated: Int
    @updated_at ?? 0
  end
end

class Document
  include Timestamped

  getter title: String

  def initialize(@title: String)
  end

  def save(now: Int)
    touch(now)
  end
end

doc = Document.new("My Doc")
doc.save(100)
puts doc.created
puts doc.updated

doc.save(200)
puts doc.created
puts doc.updated

#@ expect:
# 100
# 100
# 100
# 200
