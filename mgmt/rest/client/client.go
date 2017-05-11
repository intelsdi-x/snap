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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/v1"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
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
	// Basic http auth username/password
	Username string
	Password string
}

// Checks validity of URL
func parseURL(url string) error {
	if !govalidator.IsURL(url) || !strings.HasPrefix(url, "http") {
		return fmt.Errorf("URL %s is not in the format of http(s)://<ip>:<port>", url)
	}
	return nil
}

type metaOp func(c *Client)

//Password is an option than can be provided to the func client.New.
func Password(p string) metaOp {
	return func(c *Client) {
		c.Password = strings.TrimSpace(p)
	}
}

//Username is an option that can be provided to the func client.New.
func Username(u string) metaOp {
	return func(c *Client) {
		c.Username = u
	}
}

//Timeout is an option that can be provided to the func client.New in order to set HTTP connection timeout.
func Timeout(t time.Duration) metaOp {
	return func(c *Client) {
		c.http.Timeout = t
	}
}

var (
	secureTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		IdleConnTimeout: time.Second,
	}
	insecureTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: time.Second,
	}
)

// New returns a pointer to a snap api client
// if ver is an empty string, v1 is used by default
func New(url, ver string, insecure bool, opts ...metaOp) (*Client, error) {
	if err := parseURL(url); err != nil {
		return nil, err
	}
	if ver == "" {
		ver = "v1"
	}
	var t *http.Transport
	if insecure {
		t = insecureTransport
	} else {
		t = secureTransport
	}
	c := &Client{
		URL:     url,
		Version: ver,

		http: &http.Client{
			Transport: t,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.prefix = url + "/" + ver
	return c, nil
}

// String returns the string representation of the content type given a content number.
func (t contentType) String() string {
	return contentTypes[t]
}

/*
   Add's auth info to request if password is set.
*/
func addAuth(req *http.Request, username, password string) {
	if password != "" {
		if username == "" {
			username = "snap"
		}
		req.SetBasicAuth(username, password)
	}
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
		req *http.Request
	)
	switch method {
	case "GET":
		req, err = http.NewRequest(method, c.prefix+path, nil)
		if err != nil {
			return nil, err
		}
		addAuth(req, c.Username, c.Password)
		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		defer rsp.Body.Close()
	case "PUT":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		req, err = http.NewRequest(method, c.prefix+path, b)
		if err != nil {
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		addAuth(req, c.Username, c.Password)
		req.Header.Add("Content-Type", ct.String())

		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		defer rsp.Body.Close()
	case "DELETE":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}

		req, err = http.NewRequest(method, c.prefix+path, b)
		if err != nil {
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		addAuth(req, c.Username, c.Password)
		req.Header.Add("Content-Type", "application/json")
		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		defer rsp.Body.Close()
	case "POST":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		req, err = http.NewRequest(method, c.prefix+path, b)
		if err != nil {
			return nil, err
		}
		addAuth(req, c.Username, c.Password)
		req.Header.Add("Content-Type", ct.String())
		rsp, err = c.http.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
				return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
			}
			return nil, fmt.Errorf("URL target is not available. %v", err)
		}
		defer rsp.Body.Close()
	}
	return httpRespToAPIResp(rsp)
}

func httpRespToAPIResp(rsp *http.Response) (*rbody.APIResponse, error) {
	if rsp.StatusCode == 401 {
		return nil, fmt.Errorf("Invalid credentials")
	}
	resp := new(rbody.APIResponse)
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	jErr := json.Unmarshal(b, resp)
	// If unmarshaling fails show first part of response to help debug
	// connection issues.
	if jErr != nil {
		bound := 1000
		var invalidResp string
		if len(b) > bound {
			invalidResp = string(b[:bound])
		} else {
			invalidResp = string(b)
		}
		return nil, fmt.Errorf("Unknown API response: %s\n\n Received: %s", jErr, invalidResp)
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
	if core.IsUri(pluginPaths[0]) {
		if _, err := url.ParseRequestURI(pluginPaths[0]); err == nil {
			req, err := http.NewRequest(
				"POST",
				c.prefix+"/plugins",
				strings.NewReader(
					fmt.Sprintf("{\"uri\": \"%s\"}", pluginPaths[0]),
				),
			)
			if err != nil {
				return nil, err
			}
			addAuth(req, c.Username, c.Password)
			req.Header.Add("Content-Type", "application/json")
			rsp, err := c.http.Do(req)
			if err != nil {
				if strings.Contains(err.Error(), "tls: oversized record") || strings.Contains(err.Error(), "malformed HTTP response") {
					return nil, fmt.Errorf("error connecting to API URI: %s. Do you have an http/https mismatch?", c.URL)
				}
				return nil, fmt.Errorf("URL target is not available. %v", err)
			}
			return httpRespToAPIResp(rsp)
		}
	}
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
		if baseName := filepath.Base(pluginPath); strings.HasPrefix(baseName, v1.TLSCertPrefix) ||
			strings.HasPrefix(baseName, v1.TLSKeyPrefix) || strings.HasPrefix(baseName, v1.TLSCACertsPrefix) {
			defer os.Remove(pluginPath)
		}
		paths = append(paths, filepath.Base(pluginPath))
	}
	// with io.Pipe the write needs to be async
	go writePluginToWriter(pw, bufins, writer, paths, errChan)

	req, err := http.NewRequest("POST", c.prefix+"/plugins", pr)
	addAuth(req, c.Username, c.Password)
	if err != nil {
		return nil, fmt.Errorf("URL target is not available. %v", err)
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
		return nil, fmt.Errorf("URL target is not available. %v", err)
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

// Passthrough for tribe request to allow use of client auth.
func (c *Client) TribeRequest() (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, err
	}
	addAuth(req, "snap", c.Password)
	rsp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
