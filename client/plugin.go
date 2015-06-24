package pulse

import (
	"encoding/json"
	"errors"
	"time"
)

type Plugin struct {
	Name            string     `json:"name"`
	Version         int        `json:"version,omitempty"`
	TypeName        string     `json:"type"`
	Status          string     `json:"status,omitempty"`
	LoadedTimestamp *time.Time `json:"loaded_timestamp,omitempty"`
	HitCount        int        `json:"hit_count,omitempty"`
	LastHit         *time.Time `json:"last_hit,omitempty"`
}

func (c *Client) LoadPlugin(p string) error {
	resp, err := c.pluginUploadRequest(p)
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

func (c *Client) GetPlugins(details bool) (loadedPlugins, runningPlugins []*Plugin, err error) {
	var resp *response
	if details {
		resp, err = c.do("GET", "/plugins?details", ContentTypeJSON)
	} else {
		resp, err = c.do("GET", "/plugins", ContentTypeJSON)
	}
	if err != nil {
		return nil, nil, err
	}
	var rs getPluginsReply
	json.Unmarshal(resp.body, &rs)
	if rs.Meta.Code != 200 {
		return nil, nil, errors.New(rs.Meta.Message)
	}
	loadedPlugins = rs.Data.LoadedPlugins
	runningPlugins = rs.Data.RunningPlugins
	return
}

type getPluginsReply struct {
	respBody
	Data getPluginsData `json:"data"`
}

type getPluginsData struct {
	LoadedPlugins  []*Plugin
	RunningPlugins []*Plugin
}
