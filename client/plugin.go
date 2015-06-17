package pulse

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/intelsdi-x/pulse/control"
)

type Plugin struct {
	Name            string    `json:"name"`
	Version         int       `json:"version,omitempty"`
	TypeName        string    `json:"type"`
	Status          string    `json:"status,omitempty"`
	LoadedTimestamp int64     `json:"loaded_timestamp,omitempty"`
	HitCount        int       `json:"hit_count,omitempty"`
	LastHit         time.Time `json:"last_hit,omitempty"`
}

func (c *Client) LoadPlugin(path string) error {
	resp, err := c.do("POST", "/plugins", []byte("{\"path\": \""+path+"\"}"))
	if err != nil {
		fmt.Println(err)
		return err
	}
	var respb respBody
	err = json.Unmarshal(resp.body, &respb)
	if err != nil {
		return err
	}
	if respb.Meta.Code != 200 {
		switch resp.header.Get("Error") {
		case control.ErrPluginAlreadyLoaded.Error():
			pname := resp.header.Get("Plugin-Name")
			pversion := resp.header.Get("Plugin-Version")
			return fmt.Errorf("Plugin path(%s) already loaded as plugin (%s v%s)", path, pname, pversion)
		}
		return errors.New(respb.Meta.Message)
	}
	return nil
}

func (c *Client) GetPlugins(details bool) (loadedPlugins, runningPlugins []*Plugin, err error) {
	var resp *response
	if details {
		resp, err = c.do("GET", "/plugins?details")
	} else {
		resp, err = c.do("GET", "/plugins")
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
