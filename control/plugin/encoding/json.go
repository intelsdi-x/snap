package encoding

import (
	"bytes"
	"encoding/json"

	"github.com/intelsdi-x/pulse/control/plugin/encrypter"
)

type jsonEncoder struct {
	e *encrypter.Encrypter
}

func NewJsonEncoder() *jsonEncoder {
	return &jsonEncoder{}
}

func (j *jsonEncoder) SetEncrypter(e *encrypter.Encrypter) {
	j.e = e
}

func (j *jsonEncoder) Encode(in interface{}) ([]byte, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	if j.e != nil {
		out, err = j.e.Encrypt(bytes.NewReader(out))
	}
	return out, err
}

func (j *jsonEncoder) Decode(in []byte, out interface{}) error {
	var err error
	if j.e != nil {
		in, err = j.e.Decrypt(bytes.NewReader(in))
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(in, out)
}
