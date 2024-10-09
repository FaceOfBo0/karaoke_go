package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Idx: 0},
		"b": {Name: "b", Scope: GlobalScope, Idx: 1},
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
