package compiler

import "testing"

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("e")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Idx: 0},
				{Name: "b", Scope: GlobalScope, Idx: 1},
				{Name: "e", Scope: LocalScope, Idx: 0},
				{Name: "d", Scope: LocalScope, Idx: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Idx: 0},
				{Name: "b", Scope: GlobalScope, Idx: 1},
				{Name: "e", Scope: LocalScope, Idx: 0},
				{Name: "f", Scope: LocalScope, Idx: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}
	}
}
func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("e")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Idx: 0},
		{Name: "b", Scope: GlobalScope, Idx: 1},
		{Name: "e", Scope: LocalScope, Idx: 0},
		{Name: "d", Scope: LocalScope, Idx: 1},
	}
	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}
}

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Idx: 0},
		"b": {Name: "b", Scope: GlobalScope, Idx: 1},
		"c": {Name: "c", Scope: LocalScope, Idx: 0},
		"d": {Name: "d", Scope: LocalScope, Idx: 1},
		"e": {Name: "e", Scope: LocalScope, Idx: 0},
		"f": {Name: "f", Scope: LocalScope, Idx: 1},
	}

	globalST := NewSymbolTable()

	a := globalST.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}
	b := globalST.Define("b")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(globalST)

	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolve(t *testing.T) {
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Idx: 0},
		{Name: "b", Scope: GlobalScope, Idx: 1},
	}

	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)

		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if sym != result {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}
