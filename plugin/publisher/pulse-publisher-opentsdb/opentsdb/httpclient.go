package opentsdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	putEndPoint     = "/api/put"
	contentTypeJson = "application/json"
	userAgent       = "pulse-publisher"
)

type HttpClient struct {
	url        string
	httpClient *http.Client
	userAgent  string
}

type Client interface {
	NewClient(url string, timeout time.Duration) *HttpClient
}

//NewClient creates an instance of HttpClient which times out at
//the givin duration.
func NewClient(url string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		url: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

func (hc *HttpClient) getUrl() string {
	u := url.URL{
		Scheme: "http",
		Host:   hc.url,
		Path:   putEndPoint,
	}
	return u.String()
}

// Post stores slides of Datapoint to OpenTSDB
func (hc *HttpClient) Post(dps []DataPoint) error {
	url := hc.getUrl()

	buf, err := json.Marshal(dps)
	if err != nil {
		return err
	}

	resp, err := hc.httpClient.Post(url, contentTypeJson, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]int
	if err := json.Unmarshal(content, &result); err != nil {
		return err
	}
	return fmt.Errorf("failed to post %d data to OpenTSDB, %d succeeded", result["failed"], result["success"])
}
