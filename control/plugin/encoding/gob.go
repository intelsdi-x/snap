package encoding

import (
	"bytes"
	"encoding/gob"

	"github.com/intelsdi-x/pulse/control/plugin/encrypter"
)

type gobEncoder struct {
	e *encrypter.Encrypter
}

func NewGobEncoder() *gobEncoder {
	return &gobEncoder{}
}

func (g *gobEncoder) SetEncrypter(e *encrypter.Encrypter) {
	g.e = e
}

func (g *gobEncoder) Encode(in interface{}) ([]byte, error) {
	buff := &bytes.Buffer{}
	enc := gob.NewEncoder(buff)
	err := enc.Encode(in)
	if err != nil {
		return nil, err
	}

	if g.e != nil {
		return g.e.Encrypt(buff)
	}

	return buff.Bytes(), err
}

func (g *gobEncoder) Decode(in []byte, out interface{}) error {
	var err error
	if g.e != nil {
		in, err = g.e.Decrypt(bytes.NewReader(in))
		if err != nil {
			return err
		}
	}
	dec := gob.NewDecoder(bytes.NewReader(in))
	err = dec.Decode(out)
	if err != nil {
		return err
	}
	return nil
}
