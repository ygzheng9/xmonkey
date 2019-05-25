package object

type Environment struct {
	outer *Environment
	store map[string]Object
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func (r *Environment) Get(name string) (Object, bool) {
	obj, ok := r.store[name]
	if !ok && r.outer != nil {
		obj, ok = r.outer.Get(name)
	}

	return obj, ok
}

func (r *Environment) Set(name string, val Object) Object {
	r.store[name] = val
	return val
}

func NewEnclosedEnv(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer

	return env
}
