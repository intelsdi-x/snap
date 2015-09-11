package encoding

import "github.com/intelsdi-x/pulse/control/plugin/encrypter"

type Encoder interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	SetEncrypter(*encrypter.Encrypter)
}
