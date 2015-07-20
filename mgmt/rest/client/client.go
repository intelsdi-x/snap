package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/intelsdi-x/pulse/mgmt/rest"
)

var (
	CompressUpload = true

	ErrUnknown     = errors.New("Unknown error calling API")
	ErrNilResponse = errors.New("Nil response from JSON unmarshalling")
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
	URL     string
	Version string

	prefix string
}

// New returns a pointer to a pulse api client
// if ver is an empty string, v1 is used by default
func New(url, ver string) *Client {
	if ver == "" {
		ver = "v1"
	}
	c := &Client{
		URL:     url,
		Version: ver,
	}
	c.prefix = url + "/" + ver
	// TODO (danielscottt): assert that path is valid and target is available
	return c
}

func (t contentType) String() string {
	return contentTypes[t]
}

/*
   do handles all interactions with Pulse's REST API.
   we use the variadic function signature so that all actions can use the same
   function, including ones which do not include a request body.
*/

func (c *Client) do(method, path string, ct contentType, body ...[]byte) (*rest.APIResponse, error) {
	var (
		rsp *http.Response
		err error
	)
	switch method {
	case "GET":
		rsp, err = http.Get(c.prefix + path)
		if err != nil {
			return nil, err
		}
	case "PUT":
		client := &http.Client{}
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
		rsp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	case "DELETE":
		client := &http.Client{}
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
		rsp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	case "POST":
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		rsp, err = http.Post(c.prefix+path, ct.String(), b)
		if err != nil {
			return nil, err
		}
	}

	return httpRespToAPIResp(rsp)
}

func httpRespToAPIResp(rsp *http.Response) (*rest.APIResponse, error) {
	resp := new(rest.APIResponse)
	b, err := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	jErr := json.Unmarshal(b, resp)
	if jErr != nil {
		return nil, err
	}
	if resp == nil {
		// Catch corner case where JSON gives no error but resp is nil
		return nil, ErrNilResponse
	}
	// Add copy of JSON response string
	resp.JSONResponse = string(b)
	return resp, nil
}

func (c *Client) pluginUploadRequest(pluginPath string) (*rest.APIResponse, error) {
	client := &http.Client{}
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bufin := bufio.NewReader(file)
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// with io.Pipe the write needs to be async
	errChan := make(chan error)
	go writePluginToWriter(pw, bufin, writer, filepath.Base(pluginPath), errChan)

	req, err := http.NewRequest("POST", c.prefix+"/plugins", pr)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	if CompressUpload {
		req.Header.Add("Plugin-Compression", "gzip")
	}
	rsp, err := client.Do(req)
	cErr := <-errChan
	if cErr != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return httpRespToAPIResp(rsp)
}

func writePluginToWriter(pw io.WriteCloser, bufin *bufio.Reader, writer *multipart.Writer, pluginPath string, errChan chan error) {
	part, err := writer.CreateFormFile("pulse-plugins", pluginPath)
	if err != nil {
		errChan <- err
		return
	}
	if CompressUpload {
		cpart := gzip.NewWriter(part)
		_, err := bufin.WriteTo(cpart)
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
		_, err := bufin.WriteTo(part)
		if err != nil {
			errChan <- err
			return
		}
	}
	err = writer.Close()
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

type response struct {
	status int
	body   []byte
	header http.Header
}

type respBody struct {
	Meta *metaResp `json:"meta"`
}

type metaResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
