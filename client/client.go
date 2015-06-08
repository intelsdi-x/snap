package pulse

import (
	"bytes"
	"io/ioutil"
	"net/http"
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
	// TODO (dpittman): assert that path is valid and target is available
	return c
}

/*
   do handles all interactions with Pulse's REST API.
   we use the variadic function signature so that all actions can use the same
   function, including ones which do not include a request body.
*/
func (c *Client) do(method, path string, body ...[]byte) (*response, error) {
	var (
		rsp *http.Response
		err error
		b   []byte
	)
	switch method {
	case "GET":
		rsp, err = http.Get(c.prefix + path)
	case "PUT":
		client := &http.Client{}
		var b *bytes.Reader
		if len(body) == 0 {
			b = bytes.NewReader([]byte{})
		} else {
			b = bytes.NewReader(body[0])
		}
		req, err := http.NewRequest("PUT", c.prefix+path, b)
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
		rsp, err = http.Post(c.prefix+path, "application/json", b)
		if err != nil {
			return nil, err
		}
	}
	resp := &response{
		status: rsp.StatusCode,
		header: rsp.Header,
	}
	b, err = ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.body = b
	return resp, nil
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
