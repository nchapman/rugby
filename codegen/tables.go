package codegen

import "github.com/nchapman/rugby/ast"

// kernelFunc describes a Rugby kernel function mapping to runtime
type kernelFunc struct {
	runtimeFunc string
	// transform customizes function name based on args (e.g., rand -> RandInt/RandFloat)
	transform func(args []ast.Expression) string
}

// Kernel function mappings - single source of truth
var kernelFuncs = map[string]kernelFunc{
	"puts":  {runtimeFunc: "runtime.Puts"},
	"print": {runtimeFunc: "runtime.Print"},
	"p":     {runtimeFunc: "runtime.P"},
	"gets":  {runtimeFunc: "runtime.Gets"},
	"exit":  {runtimeFunc: "runtime.Exit"},
	"sleep": {runtimeFunc: "runtime.Sleep"},
	"rand": {
		transform: func(args []ast.Expression) string {
			if len(args) > 0 {
				return "runtime.RandInt"
			}
			return "runtime.RandFloat"
		},
	},
	"error": {runtimeFunc: "runtime.Error"},
}

// Kernel functions that can be used as identifiers (without parens)
var noParenKernelFuncs = map[string]string{
	"gets": "runtime.Gets()",
	"rand": "runtime.RandFloat()",
}

// blockMethod describes a block method mapping
type blockMethod struct {
	runtimeFunc    string
	returnType     string // "any" or "bool"
	hasAccumulator bool   // for reduce - takes 2 params (acc, elem) instead of 1
}

// Block method mappings - single source of truth
var blockMethods = map[string]blockMethod{
	"map":    {runtimeFunc: "runtime.Map", returnType: "any"},
	"select": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"filter": {runtimeFunc: "runtime.Select", returnType: "bool"},
	"reject": {runtimeFunc: "runtime.Reject", returnType: "bool"},
	"find":   {runtimeFunc: "runtime.FindPtr", returnType: "bool"},
	"detect": {runtimeFunc: "runtime.FindPtr", returnType: "bool"},
	"reduce": {runtimeFunc: "runtime.Reduce", returnType: "any", hasAccumulator: true},
	"any?":   {runtimeFunc: "runtime.Any", returnType: "bool"},
	"all?":   {runtimeFunc: "runtime.All", returnType: "bool"},
	"none?":  {runtimeFunc: "runtime.None", returnType: "bool"},
}

// Note: to_i and to_f return (T, error) - use with ! or rescue

// MethodDef describes a standard library method
type MethodDef struct {
	RuntimeFunc string
	ReturnType  string // optional, used for type inference of the result
}

// stdLib defines the standard library methods
// ReceiverType -> MethodName -> MethodDef
var stdLib = map[string]map[string]MethodDef{
	"Array": {
		"join":      {RuntimeFunc: "runtime.Join", ReturnType: "String"},
		"flatten":   {RuntimeFunc: "runtime.Flatten", ReturnType: "Array"},
		"uniq":      {RuntimeFunc: "runtime.Uniq", ReturnType: "Array"},
		"compact":   {RuntimeFunc: "runtime.Compact", ReturnType: "Array"},
		"take":      {RuntimeFunc: "runtime.Take", ReturnType: "Array"},
		"drop":      {RuntimeFunc: "runtime.Drop", ReturnType: "Array"},
		"sort":      {RuntimeFunc: "runtime.Sort", ReturnType: "Array"},
		"sorted":    {RuntimeFunc: "runtime.Sort", ReturnType: "Array"},
		"shuffle":   {RuntimeFunc: "runtime.Shuffle", ReturnType: "Array"},
		"sample":    {RuntimeFunc: "runtime.Sample", ReturnType: ""}, // T
		"first":     {RuntimeFunc: "runtime.First", ReturnType: ""},  // T
		"last":      {RuntimeFunc: "runtime.Last", ReturnType: ""},   // T
		"reverse":   {RuntimeFunc: "runtime.Reversed", ReturnType: "Array"},
		"reversed":  {RuntimeFunc: "runtime.Reversed", ReturnType: "Array"},
		"rotate":    {RuntimeFunc: "runtime.Rotate", ReturnType: "Array"},
		"include?":  {RuntimeFunc: "runtime.Contains", ReturnType: "Bool"},
		"contains?": {RuntimeFunc: "runtime.Contains", ReturnType: "Bool"},
	},
	"Map": {
		"keys":      {RuntimeFunc: "runtime.Keys", ReturnType: "Array"},
		"values":    {RuntimeFunc: "runtime.Values", ReturnType: "Array"},
		"has_key?":  {RuntimeFunc: "runtime.MapHasKey", ReturnType: "Bool"},
		"key?":      {RuntimeFunc: "runtime.MapHasKey", ReturnType: "Bool"},
		"contains?": {RuntimeFunc: "runtime.MapHasKey", ReturnType: "Bool"},
		"delete":    {RuntimeFunc: "runtime.MapDelete", ReturnType: ""}, // V
		"clear":     {RuntimeFunc: "runtime.MapClear", ReturnType: ""},
		"invert":    {RuntimeFunc: "runtime.MapInvert", ReturnType: "Map"},
		"merge":     {RuntimeFunc: "runtime.Merge", ReturnType: "Map"},
		"fetch":     {RuntimeFunc: "runtime.Fetch", ReturnType: ""},  // V
		"get":       {RuntimeFunc: "runtime.MapGet", ReturnType: ""}, // V? (optional)
		"length":    {RuntimeFunc: "runtime.MapLength", ReturnType: "Int"},
		"size":      {RuntimeFunc: "runtime.MapLength", ReturnType: "Int"},
		"empty?":    {RuntimeFunc: "runtime.MapEmpty", ReturnType: "Bool"},
	},
	"String": {
		"split":       {RuntimeFunc: "runtime.Split", ReturnType: "Array"},
		"strip":       {RuntimeFunc: "runtime.Strip", ReturnType: "String"},
		"lstrip":      {RuntimeFunc: "runtime.Lstrip", ReturnType: "String"},
		"rstrip":      {RuntimeFunc: "runtime.Rstrip", ReturnType: "String"},
		"upcase":      {RuntimeFunc: "runtime.Upcase", ReturnType: "String"},
		"downcase":    {RuntimeFunc: "runtime.Downcase", ReturnType: "String"},
		"capitalize":  {RuntimeFunc: "runtime.Capitalize", ReturnType: "String"},
		"include?":    {RuntimeFunc: "runtime.StringContains", ReturnType: "Bool"},
		"contains?":   {RuntimeFunc: "runtime.StringContains", ReturnType: "Bool"},
		"start_with?": {RuntimeFunc: "runtime.StringStartsWith", ReturnType: "Bool"},
		"end_with?":   {RuntimeFunc: "runtime.StringEndsWith", ReturnType: "Bool"},
		"empty?":      {RuntimeFunc: "runtime.StringEmpty", ReturnType: "Bool"},
		"gsub":        {RuntimeFunc: "runtime.Replace", ReturnType: "String"},
		"replace":     {RuntimeFunc: "runtime.Replace", ReturnType: "String"},
		"reverse":     {RuntimeFunc: "runtime.StringReverse", ReturnType: "String"},
		"chars":       {RuntimeFunc: "runtime.Chars", ReturnType: "Array"},
		"bytes":       {RuntimeFunc: "runtime.StringBytes", ReturnType: "[]byte"},
		"lines":       {RuntimeFunc: "runtime.StringLines", ReturnType: "Array"},
		"length":      {RuntimeFunc: "runtime.CharLength", ReturnType: "Int"},
		"to_i":        {RuntimeFunc: "runtime.StringToInt", ReturnType: "(Int, error)"},
		"to_f":        {RuntimeFunc: "runtime.StringToFloat", ReturnType: "(Float, error)"},
	},
	"Int": {
		"even?":     {RuntimeFunc: "runtime.Even", ReturnType: "Bool"},
		"odd?":      {RuntimeFunc: "runtime.Odd", ReturnType: "Bool"},
		"abs":       {RuntimeFunc: "runtime.Abs", ReturnType: "Int"},
		"clamp":     {RuntimeFunc: "runtime.Clamp", ReturnType: "Int"},
		"positive?": {RuntimeFunc: "runtime.Positive", ReturnType: "Bool"},
		"negative?": {RuntimeFunc: "runtime.Negative", ReturnType: "Bool"},
		"zero?":     {RuntimeFunc: "runtime.Zero", ReturnType: "Bool"},
		"to_s":      {RuntimeFunc: "runtime.IntToString", ReturnType: "String"},
		"to_f":      {RuntimeFunc: "runtime.IntToFloat", ReturnType: "Float"},
	},
	"Float": {
		"floor":     {RuntimeFunc: "runtime.Floor", ReturnType: "Int"},
		"ceil":      {RuntimeFunc: "runtime.Ceil", ReturnType: "Int"},
		"round":     {RuntimeFunc: "runtime.Round", ReturnType: "Int"},
		"truncate":  {RuntimeFunc: "runtime.Truncate", ReturnType: "Int"},
		"abs":       {RuntimeFunc: "runtime.FloatAbs", ReturnType: "Float"},
		"positive?": {RuntimeFunc: "runtime.FloatPositive", ReturnType: "Bool"},
		"negative?": {RuntimeFunc: "runtime.FloatNegative", ReturnType: "Bool"},
		"zero?":     {RuntimeFunc: "runtime.FloatZero", ReturnType: "Bool"},
		"to_s":      {RuntimeFunc: "runtime.FloatToString", ReturnType: "String"},
		"to_i":      {RuntimeFunc: "runtime.FloatToInt", ReturnType: "Int"},
	},
	"Math": { // Global Math module calls (Math.sqrt -> runtime.Sqrt)
		"sqrt":  {RuntimeFunc: "runtime.Sqrt", ReturnType: "Float"},
		"pow":   {RuntimeFunc: "runtime.Pow", ReturnType: "Float"},
		"ceil":  {RuntimeFunc: "runtime.Ceil", ReturnType: "Int"},
		"floor": {RuntimeFunc: "runtime.Floor", ReturnType: "Int"},
		"round": {RuntimeFunc: "runtime.Round", ReturnType: "Int"},
	},
}

// uniqueMethods maps method names to their runtime function if they are unique across the stdlib
// This helps when type inference fails.
var uniqueMethods = make(map[string]MethodDef)

// Value types that need runtime.OptionalT wrapper for optional types
var valueTypes = map[string]bool{
	"Int": true, "Int8": true, "Int16": true, "Int32": true, "Int64": true,
	"UInt": true, "UInt8": true, "UInt16": true, "UInt32": true, "UInt64": true,
	"Byte":  true, // alias for UInt8
	"Float": true, "Float32": true,
	"Bool": true, "String": true, "Rune": true,
}

func init() {
	// Populate uniqueMethods (methods that only appear in one type)
	counts := make(map[string]int)
	for _, methods := range stdLib {
		for name := range methods {
			counts[name]++
		}
	}
	// Exclude size/length from uniqueMethods since they apply to many types
	// (arrays, maps, ranges, strings) and need type-specific handling
	excludeFromUnique := map[string]bool{
		"size":   true,
		"length": true,
	}
	for _, methods := range stdLib {
		for name, def := range methods {
			if counts[name] == 1 && !excludeFromUnique[name] {
				uniqueMethods[name] = def
			}
		}
	}
}
