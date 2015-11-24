/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package encoding

import (
	"bytes"
	"encoding/gob"

	"github.com/intelsdi-x/snap/control/plugin/encrypter"
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
