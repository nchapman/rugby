// Package test provides assertion functions for Rugby tests.
// It provides two interfaces:
//   - assert: Assertions that record failures but continue execution
//   - require: Assertions that stop test execution immediately on failure
//
// Usage in Rugby tests:
//
//	describe "MyFeature" do
//	  it "works correctly" do |t|
//	    assert.equal(t, expected, actual)
//	    require.not_nil(t, value)
//	  end
//	end
package test

// Package-level assertion instances for convenient access.
// Rugby tests can use these directly: assert.equal(t, ...)
var (
	// Assertions that continue on failure (calls t.Error)
	AssertInstance = Assert{}

	// Assertions that stop on failure (calls t.Fatal)
	RequireInstance = Require{}
)
