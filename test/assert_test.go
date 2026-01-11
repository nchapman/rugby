package test

import (
	"errors"
	"testing"
)

// mockT is a mock testing.TB for testing assertions
type mockT struct {
	testing.TB
	failed      bool
	fatalCalled bool
}

func (m *mockT) Helper() {}

func (m *mockT) Errorf(format string, args ...any) {
	m.failed = true
}

func (m *mockT) Fatalf(format string, args ...any) {
	m.failed = true
	m.fatalCalled = true
}

func TestAssertEqual(t *testing.T) {
	t.Run("passes for equal values", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Equal(mt, 42, 42)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails for different values", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Equal(mt, 42, 43)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})

	t.Run("works with slices", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Equal(mt, []int{1, 2, 3}, []int{1, 2, 3})
		if mt.failed {
			t.Error("expected assertion to pass for equal slices")
		}
	})

	t.Run("fails for different slices", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Equal(mt, []int{1, 2, 3}, []int{1, 2, 4})
		if !mt.failed {
			t.Error("expected assertion to fail for different slices")
		}
	})
}

func TestAssertNotEqual(t *testing.T) {
	t.Run("passes for different values", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NotEqual(mt, 42, 43)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails for equal values", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NotEqual(mt, 42, 42)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertTrue(t *testing.T) {
	t.Run("passes for true", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.True(mt, true)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails for false", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.True(mt, false)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertFalse(t *testing.T) {
	t.Run("passes for false", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.False(mt, false)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails for true", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.False(mt, true)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertNil(t *testing.T) {
	t.Run("passes for nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Nil(mt, nil)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("passes for nil pointer", func(t *testing.T) {
		mt := &mockT{}
		var p *int
		AssertInstance.Nil(mt, p)
		if mt.failed {
			t.Error("expected assertion to pass for nil pointer")
		}
	})

	t.Run("fails for non-nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Nil(mt, 42)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertNotNil(t *testing.T) {
	t.Run("passes for non-nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NotNil(mt, 42)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails for nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NotNil(mt, nil)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertError(t *testing.T) {
	t.Run("passes when error is present", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Error(mt, errors.New("some error"))
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails when error is nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.Error(mt, nil)
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	t.Run("passes when error is nil", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NoError(mt, nil)
		if mt.failed {
			t.Error("expected assertion to pass")
		}
	})

	t.Run("fails when error is present", func(t *testing.T) {
		mt := &mockT{}
		AssertInstance.NoError(mt, errors.New("some error"))
		if !mt.failed {
			t.Error("expected assertion to fail")
		}
	})
}

func TestRequireEqual(t *testing.T) {
	t.Run("fails and calls fatal", func(t *testing.T) {
		mt := &mockT{}
		RequireInstance.Equal(mt, 42, 43)
		if !mt.failed || !mt.fatalCalled {
			t.Error("expected require to call Fatalf")
		}
	})
}
