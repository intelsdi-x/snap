package rbody

type Body interface {
	// These function names are rather verbose to avoid field vs function namespace collisions
	// with varied object types that use them.
	ResponseBodyMessage() string
	ResponseBodyType() string
}
