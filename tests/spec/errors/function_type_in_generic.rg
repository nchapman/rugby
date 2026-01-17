#@ compile-fail
# Bug: Function types not supported as generic type arguments
# Array<(): Int> and similar function type generics fail to parse

# Expected behavior:
# handlers : Array<(): Int> = []
# handlers << -> { 42 }

# Current: parser error "expected ')'" or "expected '>'"
handlers : Array<(): Int> = []
#~ ERROR: expected
