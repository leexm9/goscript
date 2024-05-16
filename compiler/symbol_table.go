package compiler

import "goscript/object"

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAl"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
	Type  object.Object
}

type FreeSymbol struct {
	Symbol
	Rewrite bool
}

type SymbolTable struct {
	Outer          *SymbolTable
	Store          map[string]Symbol
	NumDefinitions int
	FreeSymbols    []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	var free []Symbol
	return &SymbolTable{Store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (st *SymbolTable) Define(name string) Symbol {
	if name == "_" {
		return Symbol{Name: name}
	}
	symbol, ok := st.Store[name]
	if ok {
		return symbol
	} else {
		symbol = Symbol{Name: name, Index: st.NumDefinitions}
	}
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	st.Store[name] = symbol
	st.NumDefinitions++
	return symbol
}

func (st *SymbolTable) DefineWithType(name string, defObj object.Object) Symbol {
	if name == "_" {
		return Symbol{Name: name}
	}
	symbol, ok := st.Store[name]
	if ok && symbol.Type != nil {
		return symbol
	} else {
		symbol = Symbol{Name: name, Index: st.NumDefinitions, Type: defObj}
	}
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	st.Store[name] = symbol
	st.NumDefinitions++
	return symbol
}

func (st *SymbolTable) DeleteSymbol(name string) {
	_, ok := st.Store[name]
	if ok {
		delete(st.Store, name)
		st.NumDefinitions--
	}
	if !ok && st.Outer != nil {
		st.Outer.DeleteSymbol(name)
	}
}

func (st *SymbolTable) ResetSymbol(s Symbol) {
	st.Store[s.Name] = s
	st.NumDefinitions++
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := st.Store[name]
	if !ok && st.Outer != nil {
		symbol, ok = st.Outer.Resolve(name)
		if !ok {
			return symbol, ok
		}

		if symbol.Scope == GlobalScope || symbol.Scope == BuiltinScope || symbol.Scope == FunctionScope {
			return symbol, ok
		}

		free := st.defineFree(symbol)
		return free, true
	}
	return symbol, ok
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope, Type: &object.Builtin{}}
	st.Store[name] = symbol
	return symbol
}

func (st *SymbolTable) DefineFunctionName(name string, defObj object.Object) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope, Type: defObj}
	st.Store[name] = symbol
	return symbol
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)
	symbol := Symbol{
		Name:  original.Name,
		Index: len(st.FreeSymbols) - 1,
		Scope: FreeScope,
		Type:  original.Type,
	}
	st.Store[original.Name] = symbol
	return symbol
}
