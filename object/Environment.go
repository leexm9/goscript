package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (env *Environment) Set(name string, value Object) {
	env.store[name] = value
}
