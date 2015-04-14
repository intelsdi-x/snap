package core

type Plugin interface {
	Name() string
	Version() int
}
