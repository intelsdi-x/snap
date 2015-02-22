package ctypes

// TODO constructors for each that have typing for value (and optionally validate)

type ConfigValue interface {
	Type() string
}

type ConfigValueInt struct {
	Value int
}

func (c ConfigValueInt) Type() string {
	return "integer"
}

type ConfigValueStr struct {
	Value string
}

func (c ConfigValueStr) Type() string {
	return "string"
}

type ConfigValueFloat struct {
	Value float64
}

func (c ConfigValueFloat) Type() string {
	return "float"
}

// Returns a slice of string keywords for the types supported by ConfigValue.
func SupportedTypes() []string {
	// This is kind of a hack but keeps the definiton of types here in
	// ctypes.go. If you create a new ConfigValue type be sure and add here
	// to return the Type() response. This will cause any depedant components
	// to acknowledge and use that type.
	t := []string{
		// String
		ConfigValueStr{}.Type(),
		// Integer
		ConfigValueInt{}.Type(),
		// Float
		ConfigValueFloat{}.Type(),
	}
	return t
}
