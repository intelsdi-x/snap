package core

// TODO constructors for each that have typing for value (and optionally validate)

type ConfigValue interface {
	Type() string
}

type configValueInt struct {
	Value int
}

func (c configValueInt) Type() string {
	return "integer"
}

type configValueStr struct {
	Value string
}

func (c configValueStr) Type() string {
	return "string"
}

type configValueFloat struct {
	Value float64
}

func (c configValueFloat) Type() string {
	return "float"
}
