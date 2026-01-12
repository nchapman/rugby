package ast

import "testing"

func TestCommentGroup_Text(t *testing.T) {
	tests := []struct {
		name     string
		group    *CommentGroup
		expected string
	}{
		{
			name:     "nil group",
			group:    nil,
			expected: "",
		},
		{
			name:     "empty group",
			group:    &CommentGroup{List: []*Comment{}},
			expected: "",
		},
		{
			name: "single comment",
			group: &CommentGroup{
				List: []*Comment{{Text: "# hello"}},
			},
			expected: "hello",
		},
		{
			name: "single comment with leading space",
			group: &CommentGroup{
				List: []*Comment{{Text: "# hello world"}},
			},
			expected: "hello world",
		},
		{
			name: "single comment no space after #",
			group: &CommentGroup{
				List: []*Comment{{Text: "#hello"}},
			},
			expected: "hello",
		},
		{
			name: "multiple comments",
			group: &CommentGroup{
				List: []*Comment{
					{Text: "# line 1"},
					{Text: "# line 2"},
					{Text: "# line 3"},
				},
			},
			expected: "line 1\nline 2\nline 3",
		},
		{
			name: "multiple comments with varying spaces",
			group: &CommentGroup{
				List: []*Comment{
					{Text: "# single space"},
					{Text: "#no space"},
					{Text: "#  double space"},
				},
			},
			expected: "single space\nno space\n double space",
		},
		{
			name: "comment without # prefix",
			group: &CommentGroup{
				List: []*Comment{{Text: "no prefix"}},
			},
			expected: "no prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.Text()
			if got != tt.expected {
				t.Errorf("Text() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNilLit_String(t *testing.T) {
	n := &NilLit{}
	if n.String() != "nil" {
		t.Errorf("NilLit.String() = %q, want \"nil\"", n.String())
	}
}

func TestProgram_Node(t *testing.T) {
	p := &Program{}
	p.node()
	var _ Node = p
}

func TestImportDecl_Interfaces(t *testing.T) {
	i := &ImportDecl{Path: "fmt", Alias: "f"}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestFuncDecl_Interfaces(t *testing.T) {
	f := &FuncDecl{Name: "foo"}
	f.node()
	f.stmtNode()
	var _ Node = f
	var _ Statement = f
}

func TestExprStmt_Interfaces(t *testing.T) {
	e := &ExprStmt{Expr: &Ident{Name: "x"}}
	e.node()
	e.stmtNode()
	var _ Node = e
	var _ Statement = e
}

func TestBlockExpr_Interfaces(t *testing.T) {
	b := &BlockExpr{}
	b.node()
	b.exprNode()
	var _ Node = b
	var _ Expression = b
}

func TestCallExpr_Interfaces(t *testing.T) {
	c := &CallExpr{Func: &Ident{Name: "foo"}}
	c.node()
	c.exprNode()
	var _ Node = c
	var _ Expression = c
}

func TestSelectorExpr_Interfaces(t *testing.T) {
	s := &SelectorExpr{X: &Ident{Name: "x"}, Sel: "foo"}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestStringLit_Interfaces(t *testing.T) {
	s := &StringLit{Value: "hello"}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestInterpolatedString_Interfaces(t *testing.T) {
	i := &InterpolatedString{}
	i.node()
	i.exprNode()
	var _ Node = i
	var _ Expression = i
}

func TestIdent_Interfaces(t *testing.T) {
	i := &Ident{Name: "foo"}
	i.node()
	i.exprNode()
	var _ Node = i
	var _ Expression = i
}

func TestIntLit_Interfaces(t *testing.T) {
	i := &IntLit{Value: 42}
	i.node()
	i.exprNode()
	var _ Node = i
	var _ Expression = i
}

func TestFloatLit_Interfaces(t *testing.T) {
	f := &FloatLit{Value: 3.14}
	f.node()
	f.exprNode()
	var _ Node = f
	var _ Expression = f
}

func TestBoolLit_Interfaces(t *testing.T) {
	b := &BoolLit{Value: true}
	b.node()
	b.exprNode()
	var _ Node = b
	var _ Expression = b
}

func TestNilLit_Interfaces(t *testing.T) {
	n := &NilLit{}
	n.node()
	n.exprNode()
	var _ Node = n
	var _ Expression = n
}

func TestSymbolLit_Interfaces(t *testing.T) {
	s := &SymbolLit{Value: "ok"}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestArrayLit_Interfaces(t *testing.T) {
	a := &ArrayLit{}
	a.node()
	a.exprNode()
	var _ Node = a
	var _ Expression = a
}

func TestSplatExpr_Interfaces(t *testing.T) {
	s := &SplatExpr{Expr: &Ident{Name: "arr"}}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestDoubleSplatExpr_Interfaces(t *testing.T) {
	d := &DoubleSplatExpr{Expr: &Ident{Name: "map"}}
	d.node()
	d.exprNode()
	var _ Node = d
	var _ Expression = d
}

func TestIndexExpr_Interfaces(t *testing.T) {
	i := &IndexExpr{Left: &Ident{Name: "arr"}, Index: &IntLit{Value: 0}}
	i.node()
	i.exprNode()
	var _ Node = i
	var _ Expression = i
}

func TestMapLit_Interfaces(t *testing.T) {
	m := &MapLit{}
	m.node()
	m.exprNode()
	var _ Node = m
	var _ Expression = m
}

func TestRangeLit_Interfaces(t *testing.T) {
	r := &RangeLit{Start: &IntLit{Value: 1}, End: &IntLit{Value: 10}}
	r.node()
	r.exprNode()
	var _ Node = r
	var _ Expression = r
}

func TestBinaryExpr_Interfaces(t *testing.T) {
	b := &BinaryExpr{Left: &IntLit{Value: 1}, Op: "+", Right: &IntLit{Value: 2}}
	b.node()
	b.exprNode()
	var _ Node = b
	var _ Expression = b
}

func TestUnaryExpr_Interfaces(t *testing.T) {
	u := &UnaryExpr{Op: "-", Expr: &IntLit{Value: 1}}
	u.node()
	u.exprNode()
	var _ Node = u
	var _ Expression = u
}

func TestBangExpr_Interfaces(t *testing.T) {
	b := &BangExpr{Expr: &CallExpr{Func: &Ident{Name: "foo"}}}
	b.node()
	b.exprNode()
	var _ Node = b
	var _ Expression = b
}

func TestRescueExpr_Interfaces(t *testing.T) {
	r := &RescueExpr{Expr: &CallExpr{Func: &Ident{Name: "foo"}}}
	r.node()
	r.exprNode()
	var _ Node = r
	var _ Expression = r
}

func TestNilCoalesceExpr_Interfaces(t *testing.T) {
	n := &NilCoalesceExpr{Left: &Ident{Name: "x"}, Right: &IntLit{Value: 0}}
	n.node()
	n.exprNode()
	var _ Node = n
	var _ Expression = n
}

func TestTernaryExpr_Interfaces(t *testing.T) {
	te := &TernaryExpr{
		Condition: &BoolLit{Value: true},
		Then:      &IntLit{Value: 1},
		Else:      &IntLit{Value: 2},
	}
	te.node()
	te.exprNode()
	var _ Node = te
	var _ Expression = te
}

func TestSymbolToProcExpr_Interfaces(t *testing.T) {
	s := &SymbolToProcExpr{Method: "upcase"}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestSafeNavExpr_Interfaces(t *testing.T) {
	s := &SafeNavExpr{Receiver: &Ident{Name: "x"}, Selector: "name"}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestAssignStmt_Interfaces(t *testing.T) {
	a := &AssignStmt{Name: "x", Value: &IntLit{Value: 1}}
	a.node()
	a.stmtNode()
	var _ Node = a
	var _ Statement = a
}

func TestOrAssignStmt_Interfaces(t *testing.T) {
	o := &OrAssignStmt{Name: "x", Value: &IntLit{Value: 1}}
	o.node()
	o.stmtNode()
	var _ Node = o
	var _ Statement = o
}

func TestCompoundAssignStmt_Interfaces(t *testing.T) {
	c := &CompoundAssignStmt{Name: "x", Op: "+", Value: &IntLit{Value: 1}}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestMultiAssignStmt_Interfaces(t *testing.T) {
	m := &MultiAssignStmt{Names: []string{"a", "b"}}
	m.node()
	m.stmtNode()
	var _ Node = m
	var _ Statement = m
}

func TestIfStmt_Interfaces(t *testing.T) {
	i := &IfStmt{Cond: &BoolLit{Value: true}}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestCaseStmt_Interfaces(t *testing.T) {
	c := &CaseStmt{Subject: &Ident{Name: "x"}}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestCaseTypeStmt_Interfaces(t *testing.T) {
	c := &CaseTypeStmt{Subject: &Ident{Name: "x"}}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestWhileStmt_Interfaces(t *testing.T) {
	w := &WhileStmt{Cond: &BoolLit{Value: true}}
	w.node()
	w.stmtNode()
	var _ Node = w
	var _ Statement = w
}

func TestUntilStmt_Interfaces(t *testing.T) {
	u := &UntilStmt{Cond: &BoolLit{Value: false}}
	u.node()
	u.stmtNode()
	var _ Node = u
	var _ Statement = u
}

func TestForStmt_Interfaces(t *testing.T) {
	f := &ForStmt{Var: "i", Iterable: &Ident{Name: "items"}}
	f.node()
	f.stmtNode()
	var _ Node = f
	var _ Statement = f
}

func TestBreakStmt_Interfaces(t *testing.T) {
	b := &BreakStmt{}
	b.node()
	b.stmtNode()
	var _ Node = b
	var _ Statement = b
}

func TestNextStmt_Interfaces(t *testing.T) {
	n := &NextStmt{}
	n.node()
	n.stmtNode()
	var _ Node = n
	var _ Statement = n
}

func TestReturnStmt_Interfaces(t *testing.T) {
	r := &ReturnStmt{}
	r.node()
	r.stmtNode()
	var _ Node = r
	var _ Statement = r
}

func TestPanicStmt_Interfaces(t *testing.T) {
	p := &PanicStmt{Message: &StringLit{Value: "error"}}
	p.node()
	p.stmtNode()
	var _ Node = p
	var _ Statement = p
}

func TestDeferStmt_Interfaces(t *testing.T) {
	d := &DeferStmt{Call: &CallExpr{Func: &Ident{Name: "close"}}}
	d.node()
	d.stmtNode()
	var _ Node = d
	var _ Statement = d
}

func TestClassDecl_Interfaces(t *testing.T) {
	c := &ClassDecl{Name: "User"}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestMethodDecl_Interfaces(t *testing.T) {
	m := &MethodDecl{Name: "foo"}
	m.node()
	m.stmtNode()
	var _ Node = m
	var _ Statement = m
}

func TestInstanceVar_Interfaces(t *testing.T) {
	i := &InstanceVar{Name: "name"}
	i.node()
	i.exprNode()
	var _ Node = i
	var _ Expression = i
}

func TestInstanceVarAssign_Interfaces(t *testing.T) {
	i := &InstanceVarAssign{Name: "name", Value: &StringLit{Value: "test"}}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestInstanceVarOrAssign_Interfaces(t *testing.T) {
	i := &InstanceVarOrAssign{Name: "name", Value: &StringLit{Value: "default"}}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestInterfaceDecl_Interfaces(t *testing.T) {
	i := &InterfaceDecl{Name: "Speaker"}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestDescribeStmt_Interfaces(t *testing.T) {
	d := &DescribeStmt{Name: "test suite"}
	d.node()
	d.stmtNode()
	var _ Node = d
	var _ Statement = d
}

func TestItStmt_Interfaces(t *testing.T) {
	i := &ItStmt{Name: "should work"}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestTestStmt_Interfaces(t *testing.T) {
	ts := &TestStmt{Name: "my test"}
	ts.node()
	ts.stmtNode()
	var _ Node = ts
	var _ Statement = ts
}

func TestTableStmt_Interfaces(t *testing.T) {
	ts := &TableStmt{Name: "table test"}
	ts.node()
	ts.stmtNode()
	var _ Node = ts
	var _ Statement = ts
}

func TestBeforeStmt_Interfaces(t *testing.T) {
	b := &BeforeStmt{}
	b.node()
	b.stmtNode()
	var _ Node = b
	var _ Statement = b
}

func TestAfterStmt_Interfaces(t *testing.T) {
	a := &AfterStmt{}
	a.node()
	a.stmtNode()
	var _ Node = a
	var _ Statement = a
}

func TestAccessorDecl_Interfaces(t *testing.T) {
	a := &AccessorDecl{Kind: "getter", Name: "name", Type: "String"}
	a.node()
	a.stmtNode()
	var _ Node = a
	var _ Statement = a
}

func TestSuperExpr_Interfaces(t *testing.T) {
	s := &SuperExpr{}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestGoStmt_Interfaces(t *testing.T) {
	g := &GoStmt{Call: &CallExpr{Func: &Ident{Name: "work"}}}
	g.node()
	g.stmtNode()
	var _ Node = g
	var _ Statement = g
}

func TestChanSendStmt_Interfaces(t *testing.T) {
	c := &ChanSendStmt{Chan: &Ident{Name: "ch"}, Value: &IntLit{Value: 1}}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestSelectStmt_Interfaces(t *testing.T) {
	s := &SelectStmt{}
	s.node()
	s.stmtNode()
	var _ Node = s
	var _ Statement = s
}

func TestModuleDecl_Interfaces(t *testing.T) {
	m := &ModuleDecl{Name: "Loggable"}
	m.node()
	m.stmtNode()
	var _ Node = m
	var _ Statement = m
}

func TestIncludeStmt_Interfaces(t *testing.T) {
	i := &IncludeStmt{Module: "Loggable"}
	i.node()
	i.stmtNode()
	var _ Node = i
	var _ Statement = i
}

func TestSpawnExpr_Interfaces(t *testing.T) {
	s := &SpawnExpr{Block: &BlockExpr{}}
	s.node()
	s.exprNode()
	var _ Node = s
	var _ Expression = s
}

func TestAwaitExpr_Interfaces(t *testing.T) {
	a := &AwaitExpr{Task: &Ident{Name: "t"}}
	a.node()
	a.exprNode()
	var _ Node = a
	var _ Expression = a
}

func TestConcurrentlyStmt_Interfaces(t *testing.T) {
	c := &ConcurrentlyStmt{ScopeVar: "scope"}
	c.node()
	c.stmtNode()
	var _ Node = c
	var _ Statement = c
}

func TestComment_Fields(t *testing.T) {
	c := &Comment{Text: "# test", Line: 5, Col: 10}
	if c.Text != "# test" {
		t.Errorf("Comment.Text = %q, want \"# test\"", c.Text)
	}
	if c.Line != 5 {
		t.Errorf("Comment.Line = %d, want 5", c.Line)
	}
	if c.Col != 10 {
		t.Errorf("Comment.Col = %d, want 10", c.Col)
	}
}

func TestParam_Fields(t *testing.T) {
	p := &Param{Name: "x", Type: "Int"}
	if p.Name != "x" {
		t.Errorf("Param.Name = %q, want \"x\"", p.Name)
	}
	if p.Type != "Int" {
		t.Errorf("Param.Type = %q, want \"Int\"", p.Type)
	}
}

func TestFieldDecl_Fields(t *testing.T) {
	f := &FieldDecl{Name: "name", Type: "String"}
	if f.Name != "name" {
		t.Errorf("FieldDecl.Name = %q, want \"name\"", f.Name)
	}
	if f.Type != "String" {
		t.Errorf("FieldDecl.Type = %q, want \"String\"", f.Type)
	}
}

func TestMethodSig_Fields(t *testing.T) {
	m := &MethodSig{
		Name:        "speak",
		Params:      []*Param{{Name: "msg", Type: "String"}},
		ReturnTypes: []string{"String"},
	}
	if m.Name != "speak" {
		t.Errorf("MethodSig.Name = %q, want \"speak\"", m.Name)
	}
	if len(m.Params) != 1 {
		t.Errorf("len(MethodSig.Params) = %d, want 1", len(m.Params))
	}
	if len(m.ReturnTypes) != 1 {
		t.Errorf("len(MethodSig.ReturnTypes) = %d, want 1", len(m.ReturnTypes))
	}
}

func TestElseIfClause_Fields(t *testing.T) {
	e := ElseIfClause{
		Cond: &BoolLit{Value: true},
		Body: []Statement{&ReturnStmt{}},
	}
	if e.Cond == nil {
		t.Error("ElseIfClause.Cond is nil")
	}
	if len(e.Body) != 1 {
		t.Errorf("len(ElseIfClause.Body) = %d, want 1", len(e.Body))
	}
}

func TestWhenClause_Fields(t *testing.T) {
	w := WhenClause{
		Values: []Expression{&IntLit{Value: 1}, &IntLit{Value: 2}},
		Body:   []Statement{&ReturnStmt{}},
	}
	if len(w.Values) != 2 {
		t.Errorf("len(WhenClause.Values) = %d, want 2", len(w.Values))
	}
	if len(w.Body) != 1 {
		t.Errorf("len(WhenClause.Body) = %d, want 1", len(w.Body))
	}
}

func TestTypeWhenClause_Fields(t *testing.T) {
	tw := TypeWhenClause{
		Type: "String",
		Body: []Statement{&ReturnStmt{}},
	}
	if tw.Type != "String" {
		t.Errorf("TypeWhenClause.Type = %q, want \"String\"", tw.Type)
	}
	if len(tw.Body) != 1 {
		t.Errorf("len(TypeWhenClause.Body) = %d, want 1", len(tw.Body))
	}
}

func TestSelectCase_Fields(t *testing.T) {
	sc := SelectCase{
		AssignName: "val",
		Chan:       &Ident{Name: "ch"},
		IsSend:     false,
		Value:      nil,
		Body:       []Statement{&ExprStmt{Expr: &Ident{Name: "x"}}},
	}
	if sc.AssignName != "val" {
		t.Errorf("SelectCase.AssignName = %q, want \"val\"", sc.AssignName)
	}
	if sc.Chan == nil {
		t.Error("SelectCase.Chan is nil")
	}
	if sc.IsSend {
		t.Error("SelectCase.IsSend = true, want false")
	}
	if sc.Value != nil {
		t.Error("SelectCase.Value should be nil for receive")
	}
	if len(sc.Body) != 1 {
		t.Errorf("len(SelectCase.Body) = %d, want 1", len(sc.Body))
	}
}

func TestMapEntry_Fields(t *testing.T) {
	entry := MapEntry{
		Key:   &StringLit{Value: "name"},
		Value: &StringLit{Value: "Alice"},
		Splat: nil,
	}
	if entry.Key == nil {
		t.Error("MapEntry.Key is nil")
	}
	if entry.Value == nil {
		t.Error("MapEntry.Value is nil")
	}
	if entry.Splat != nil {
		t.Error("MapEntry.Splat is not nil")
	}

	splatEntry := MapEntry{
		Key:   nil,
		Value: nil,
		Splat: &Ident{Name: "defaults"},
	}
	if splatEntry.Key != nil {
		t.Error("splat MapEntry.Key should be nil")
	}
	if splatEntry.Value != nil {
		t.Error("splat MapEntry.Value should be nil")
	}
	if splatEntry.Splat == nil {
		t.Error("splat MapEntry.Splat is nil")
	}
}
