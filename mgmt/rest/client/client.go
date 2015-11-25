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

package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
)

var (
	CompressUpload = true

	ErrUnknown     = errors.New("Unknown error calling API")
	ErrNilResponse = errors.New("Nil response from JSON unmarshalling")
	ErrDirNotFile  = errors.New("Provided plugin path is a directory not file")
)

const (
	ContentTypeJSON contentType = iota
	ContentTypeBinary
)

type contentType int

var (
	contentTypes = [...]string{
		"application/json",
		"application/octet-stream",
	}
)

type Client struct {
	// URL specifies HTTP API request uniform resource locator.
	URL string
	// Version specifies the version of a HTTP client.
	Version string
	// http is a pointer to a net/http client.
	http *http.Client
	// prefix is the string concatenation of a request URL, forward slash
	// and the request client version.
	prefix string
}

// New returns a pointer to a snap api client
// if ver is an empty string, v1 is used by default
func New(url, ver string, insecure bool) *Client {
	if ver == "" {
		ver = "v1"
	}
	c := &Client{
		URL:     url,
		Version: ver,

		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
			},
		},
	}
	c.prefix = url + "/" + ver
	// TODO (danielscottt): assert that path is valid and target is available
	return c
}

// String returns the string representation of the content type given a content number.
func (t contentType) String() string {
	return contentTypes[t]
}

/*
   do handles all interactions with snap's REST API.
   we use the variadic function signature so that all actions can use the same
   function, including ones which do not include a request body.
*/

func (c *Client) do(method, path string, ct contentType, body ...[]byte) (*rbody.APIResponse, error) {
	var (
		rsp *http.Response
		err error
	)
	switch method {
	case "GET":
		rsp, err = c.http.Get(c.prefix + path)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, err
		}
	case "PUT":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		req, err := http.NewRequest("PUT", c.prefix+path, b)
		req.Header.Add("Content-Type", ct.String())
		if err != nil {
			return nil, err
		}
		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, err
		}
	case "DELETE":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		req, err := http.NewRequest("DELETE", c.prefix+path, b)
		req.Header.Add("Content-Type", "application/json")
		if err != nil {
			return nil, err
		}
		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, err
		}
	case "POST":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		rsp, err = c.http.Post(c.prefix+path, ct.String(), b)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, err
		}
	}

	return httpRespToAPIResp(rsp)
}

func httpRespToAPIResp(rsp *http.Response) (*rbody.APIResponse, error) {
	resp := new(rbody.APIResponse)
	b, err := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		return nil, err
	}
	jErr := json.Unmarshal(b, resp)
	if jErr != nil {
		return nil, jErr
	}
	if resp == nil {
		// Catch corner case where JSON gives no error but resp is nil
		return nil, ErrNilResponse
	}
	// Add copy of JSON response string
	resp.JSONResponse = string(b)
	return resp, nil
}

func (c *Client) pluginUploadRequest(pluginPaths []string) (*rbody.APIResponse, error) {
	errChan := make(chan error)
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	var bufins []*bufio.Reader
	var paths []string

	for i, pluginPath := range pluginPaths {
		path, err := os.Stat(pluginPaths[i])
		if err != nil {
			return nil, err
		}
		if path.IsDir() {
			err := ErrDirNotFile
			return nil, err
		}
		file, err := os.Open(pluginPaths[i])
		if err != nil {
			return nil, err
		}
		defer file.Close()
		bufin := bufio.NewReader(file)
		bufins = append(bufins, bufin)

		paths = append(paths, filepath.Base(pluginPath))
	}
	// with io.Pipe the write needs to be async
	go writePluginToWriter(pw, bufins, writer, paths, errChan)

	req, err := http.NewRequest("POST", c.prefix+"/plugins", pr)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	if CompressUpload {
		req.Header.Add("Plugin-Compression", "gzip")
	}
	rsp, err := c.http.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
			return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
		}
		return nil, err
	}
	cErr := <-errChan
	if cErr != nil {
		return nil, cErr
	}
	return httpRespToAPIResp(rsp)
}

func writePluginToWriter(pw io.WriteCloser, bufin []*bufio.Reader, writer *multipart.Writer, pluginPaths []string, errChan chan error) {
	for i, pluginPath := range pluginPaths {
		part, err := writer.CreateFormFile("snap-plugins", pluginPath)
		if err != nil {
			errChan <- err
			return
		}
		if CompressUpload {
			cpart := gzip.NewWriter(part)
			_, err := bufin[i].WriteTo(cpart)
			if err != nil {
				errChan <- err
				return
			}
			err = cpart.Close()
			if err != nil {
				errChan <- err
				return
			}
		} else {
			_, err := bufin[i].WriteTo(part)
			if err != nil {
				errChan <- err
				return
			}
		}
	}
	err := writer.Close()
	if err != nil {
		errChan <- err
		return
	}
	err = pw.Close()
	if err != nil {
		errChan <- err
		return
	}
	errChan <- nil
}
