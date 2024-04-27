package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

type EnvObject struct {
	Value Object
	Depth int
}

func (eb *EnvObject) GetValue() Object {
	return eb.Value
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (env *Environment) Get(name string) (EnvObject, bool) {
	obj, depth, ok := env.get(name, 0)
	return EnvObject{Value: obj, Depth: depth}, ok
}

func (env *Environment) get(name string, depth int) (Object, int, bool) {
	if name == "_" {
		return nil, 0, false
	}
	obj, ok := env.store[name]
	if !ok && env.outer != nil {
		obj, depth, ok = env.outer.get(name, depth)
		depth++
	}
	return obj, depth, ok
}

func (env *Environment) Set(name string, value Object) (Object, *Error) {
	if name == "_" {
		return value, nil
	}
	obj, _, ok := env.get(name, 0)
	if ok && value.Type() != FUNCTION_OBJ && obj.Type() != value.Type() {
		return obj, NewError("cannot use '%s' (untyped %s constant) as %s value in assignment", value, value.Type(), obj.Type())
	}
	env.store[name] = value
	return obj, nil
}

func (env *Environment) SetWithDepth(name string, value Object, depth int) (Object, *Error) {
	if name == "_" {
		return value, nil
	}
	if depth == 0 {
		obj, _, ok := env.get(name, 0)
		if ok && value.Type() != FUNCTION_OBJ && obj.Type() != value.Type() {
			return obj, NewError("cannot use '%s' (untyped %s constant) as %s value in assignment", value, value.Type(), obj.Type())
		}
		env.store[name] = value
		return obj, nil
	} else {
		return env.outer.SetWithDepth(name, value, depth-1)
	}
}

func (env *Environment) GetStore() map[string]Object {
	return env.store
}
