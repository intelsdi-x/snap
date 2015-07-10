package perror

type PulseError interface {
	error
	Fields() map[string]interface{}
	SetFields(map[string]interface{})
}

type pulseError struct {
	err    error
	fields map[string]interface{}
}

func New(e error) *pulseError {
	// Catch someone trying to wrap a pe around a pe.
	// We throw a panic to make them fix this.
	if _, ok := e.(PulseError); ok {
		panic("You are trying to wrap a pulseError around a PulseError. Don't do this.")
	}

	return &pulseError{err: e, fields: make(map[string]interface{})}
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
