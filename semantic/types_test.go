package semantic

import "testing"

func TestTypeString(t *testing.T) {
	tests := []struct {
		typ  *Type
		want string
	}{
		{TypeIntVal, "Int"},
		{TypeInt64Val, "Int64"},
		{TypeFloatVal, "Float"},
		{TypeBoolVal, "Bool"},
		{TypeStringVal, "String"},
		{TypeBytesVal, "Bytes"},
		{TypeNilVal, "nil"},
		{TypeAnyVal, "any"},
		{TypeErrorVal, "error"},
		{TypeRangeVal, "Range"},
		{TypeUnknownVal, "unknown"},
		{NewArrayType(TypeIntVal), "Array[Int]"},
		{NewArrayType(TypeStringVal), "Array[String]"},
		{NewMapType(TypeStringVal, TypeIntVal), "Map[String, Int]"},
		{NewChanType(TypeIntVal), "Chan[Int]"},
		{NewTaskType(TypeStringVal), "Task[String]"},
		{NewOptionalType(TypeIntVal), "Int?"},
		{NewOptionalType(TypeStringVal), "String?"},
		{NewTupleType(TypeIntVal, TypeBoolVal), "(Int, Bool)"},
		{NewTupleType(TypeStringVal, TypeErrorVal), "(String, error)"},
		{NewClassType("User"), "User"},
		{NewInterfaceType("Reader"), "Reader"},
		{NewFuncType([]*Type{TypeIntVal}, []*Type{TypeBoolVal}), "(Int) -> Bool"},
		{NewFuncType([]*Type{TypeIntVal, TypeIntVal}, []*Type{TypeIntVal}), "(Int, Int) -> Int"},
		{NewFuncType([]*Type{}, []*Type{}), "() -> ()"},
		{nil, "unknown"},
	}

	for _, tt := range tests {
		got := tt.typ.String()
		if got != tt.want {
			t.Errorf("Type.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestTypeEquals(t *testing.T) {
	tests := []struct {
		a, b *Type
		want bool
	}{
		{TypeIntVal, TypeIntVal, true},
		{TypeIntVal, TypeFloatVal, false},
		{NewArrayType(TypeIntVal), NewArrayType(TypeIntVal), true},
		{NewArrayType(TypeIntVal), NewArrayType(TypeStringVal), false},
		{NewMapType(TypeStringVal, TypeIntVal), NewMapType(TypeStringVal, TypeIntVal), true},
		{NewMapType(TypeStringVal, TypeIntVal), NewMapType(TypeIntVal, TypeIntVal), false},
		{NewOptionalType(TypeIntVal), NewOptionalType(TypeIntVal), true},
		{NewOptionalType(TypeIntVal), TypeIntVal, false},
		{NewTupleType(TypeIntVal, TypeBoolVal), NewTupleType(TypeIntVal, TypeBoolVal), true},
		{NewTupleType(TypeIntVal, TypeBoolVal), NewTupleType(TypeIntVal, TypeIntVal), false},
		{NewClassType("User"), NewClassType("User"), true},
		{NewClassType("User"), NewClassType("Admin"), false},
		{nil, nil, true},
		{TypeIntVal, nil, false},
	}

	for _, tt := range tests {
		got := tt.a.Equals(tt.b)
		if got != tt.want {
			t.Errorf("%v.Equals(%v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTypeIsOptional(t *testing.T) {
	if !NewOptionalType(TypeIntVal).IsOptional() {
		t.Error("Optional type should return true for IsOptional()")
	}
	if TypeIntVal.IsOptional() {
		t.Error("Int should return false for IsOptional()")
	}
}

func TestTypeUnwrap(t *testing.T) {
	opt := NewOptionalType(TypeIntVal)
	if !opt.Unwrap().Equals(TypeIntVal) {
		t.Errorf("Unwrap() = %v, want Int", opt.Unwrap())
	}

	if !TypeIntVal.Unwrap().Equals(TypeIntVal) {
		t.Error("Unwrap() on non-optional should return itself")
	}
}
