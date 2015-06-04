package pulse

import (
	"encoding/json"
	"errors"
)

func (c *Client) LoadPlugin(path string) error {
	resp, err := c.do("POST", "/plugins", []byte("{\"path\": \""+path+"\"}"))
	if err != nil {
		return err
	}
	var respb respBody
	json.Unmarshal(resp.body, &respb)
	if respb.Meta.Code != 200 {
		return errors.New(respb.Meta.Message)
	}
	return nil
}
