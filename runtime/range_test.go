package runtime

import (
	"reflect"
	"testing"
)

func TestRangeEachInclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: false}
	var collected []int
	RangeEach(r, func(i int) bool {
		collected = append(collected, i)
		return true
	})
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(collected, expected) {
		t.Errorf("RangeEach inclusive: got %v, want %v", collected, expected)
	}
}

func TestRangeEachExclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: true}
	var collected []int
	RangeEach(r, func(i int) bool {
		collected = append(collected, i)
		return true
	})
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(collected, expected) {
		t.Errorf("RangeEach exclusive: got %v, want %v", collected, expected)
	}
}

func TestRangeEachWithBreak(t *testing.T) {
	r := Range{Start: 1, End: 10, Exclusive: false}
	var collected []int
	RangeEach(r, func(i int) bool {
		collected = append(collected, i)
		return i < 3
	})
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(collected, expected) {
		t.Errorf("RangeEach with break: got %v, want %v", collected, expected)
	}
}

func TestRangeEachEmpty(t *testing.T) {
	r := Range{Start: 5, End: 1, Exclusive: false}
	var count int
	RangeEach(r, func(i int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("RangeEach empty: got %d iterations, want 0", count)
	}
}

func TestRangeContainsInclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: false}

	tests := []struct {
		n        int
		expected bool
	}{
		{0, false},
		{1, true},
		{3, true},
		{5, true},
		{6, false},
	}

	for _, tt := range tests {
		got := RangeContains(r, tt.n)
		if got != tt.expected {
			t.Errorf("RangeContains(%d) inclusive: got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestRangeContainsExclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: true}

	tests := []struct {
		n        int
		expected bool
	}{
		{0, false},
		{1, true},
		{3, true},
		{4, true},
		{5, false}, // exclusive end
		{6, false},
	}

	for _, tt := range tests {
		got := RangeContains(r, tt.n)
		if got != tt.expected {
			t.Errorf("RangeContains(%d) exclusive: got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestRangeToArrayInclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: false}
	result := RangeToArray(r)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RangeToArray inclusive: got %v, want %v", result, expected)
	}
}

func TestRangeToArrayExclusive(t *testing.T) {
	r := Range{Start: 1, End: 5, Exclusive: true}
	result := RangeToArray(r)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("RangeToArray exclusive: got %v, want %v", result, expected)
	}
}

func TestRangeToArrayEmpty(t *testing.T) {
	r := Range{Start: 5, End: 1, Exclusive: false}
	result := RangeToArray(r)
	if len(result) != 0 {
		t.Errorf("RangeToArray empty: got %v, want empty", result)
	}
}

func TestRangeSizeInclusive(t *testing.T) {
	tests := []struct {
		start, end int
		expected   int
	}{
		{1, 5, 5},  // 1, 2, 3, 4, 5
		{0, 0, 1},  // just 0
		{5, 1, 0},  // empty
		{-2, 2, 5}, // -2, -1, 0, 1, 2
	}

	for _, tt := range tests {
		r := Range{Start: tt.start, End: tt.end, Exclusive: false}
		got := RangeSize(r)
		if got != tt.expected {
			t.Errorf("RangeSize(%d..%d): got %d, want %d", tt.start, tt.end, got, tt.expected)
		}
	}
}

func TestRangeSizeExclusive(t *testing.T) {
	tests := []struct {
		start, end int
		expected   int
	}{
		{1, 5, 4},  // 1, 2, 3, 4
		{0, 1, 1},  // just 0
		{0, 0, 0},  // empty
		{5, 1, 0},  // empty
		{-2, 2, 4}, // -2, -1, 0, 1
	}

	for _, tt := range tests {
		r := Range{Start: tt.start, End: tt.end, Exclusive: true}
		got := RangeSize(r)
		if got != tt.expected {
			t.Errorf("RangeSize(%d...%d): got %d, want %d", tt.start, tt.end, got, tt.expected)
		}
	}
}
