package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Idx   int
}

type SymbolTable struct {
	store   map[string]Symbol
	numDefs int
	Outer   *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (st *SymbolTable) Define(name string) Symbol {
	sym := Symbol{Scope: GlobalScope, Name: name, Idx: st.numDefs}
	st.store[name] = sym
	st.numDefs++
	return sym
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	result, ok := st.store[name]
	return result, ok
}
