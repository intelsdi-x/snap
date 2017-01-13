/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/urfave/negroni"
)

func respond(code int, b rbody.Body, w http.ResponseWriter) {
	resp := &rbody.APIResponse{
		Meta: &rbody.APIResponseMeta{
			Code:    code,
			Message: b.ResponseBodyMessage(),
			Type:    b.ResponseBodyType(),
			Version: APIVersion,
		},
		Body: b,
	}
	if !w.(negroni.ResponseWriter).Written() {
		w.WriteHeader(code)
	}

	j, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic(err)
	}
	j = bytes.Replace(j, []byte("\\u0026"), []byte("&"), -1)
	fmt.Fprint(w, string(j))
}
