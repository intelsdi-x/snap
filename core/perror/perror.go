package perror

type PulseError interface {
	error
	Fields() map[string]interface{}
	SetFields(map[string]interface{})
}

type Fields map[string]interface{}

type pulseError struct {
	err    error
	fields Fields
}

// New returns an initialized PulseError.
// The variadic signature allows fields to optionally
// be added at construction.
func New(e error, fields ...map[string]interface{}) *pulseError {
	// Catch someone trying to wrap a pe around a pe.
	// We throw a panic to make them fix this.
	if _, ok := e.(PulseError); ok {
		panic("You are trying to wrap a pulseError around a PulseError. Don't do this.")
	}

	p := &pulseError{
		err:    e,
		fields: make(map[string]interface{}),
	}

	// insert fields into new PulseError
	for _, f := range fields {
		for k, v := range f {
			p.fields[k] = v
		}
	}

	return p
}

func (p *pulseError) SetFields(f map[string]interface{}) {
	p.fields = f
}

func (p *pulseError) Fields() map[string]interface{} {
	return p.fields
}

func (p *pulseError) Error() string {
	return p.err.Error()
}

func (p *pulseError) String() string {
	return p.Error()
}
