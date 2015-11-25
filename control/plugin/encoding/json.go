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
	"encoding/json"

	"github.com/intelsdi-x/snap/control/plugin/encrypter"
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
