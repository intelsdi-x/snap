// PluginManger manages loading, unloading, and swapping
// of plugins
package control

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	// LoadedPlugin States
	DetectedState pluginState = "detected"
	LoadingState  pluginState = "loading"
	LoadedState   pluginState = "loaded"
	UnloadedState pluginState = "unloaded"
)

type pluginState string

type loadedPlugins []*LoadedPlugin

type LoadedPlugin struct {
	Meta       plugin.PluginMeta
	Path       string
	Type       plugin.PluginType
	State      pluginState
	Token      string
	LoadedTime time.Time
}

type pluginManager struct {
	LoadedPlugins loadedPlugins
	Started       bool

	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func PluginManager() *pluginManager {
	p := new(pluginManager)
	return p
}

func (p *pluginManager) Start() {
	p.Started = true
}

func (p *pluginManager) Stop() {
	p.Started = false
}

func (p *pluginManager) GenerateArgs(daemon bool) plugin.Arg {
	a := plugin.Arg{
		ControlPubKey: p.pubKey,
		PluginLogPath: "/tmp",
		RunAsDaemon:   daemon,
	}
	return a
}

func (p *pluginManager) LoadPlugin(path string) error {
	if !p.Started {
		return errors.New("Must start pluginManager before calling LoadPlugin()")
	}

	if err := load(p, path); err != nil {
		return err
	}

	return nil
}

func load(p *pluginManager, path string) error {
	log.Printf("Attempting to load: %s\v", path)
	lPlugin := new(LoadedPlugin)
	lPlugin.Path = path
	lPlugin.State = DetectedState

	ePlugin, err := plugin.NewExecutablePlugin(p, lPlugin.Path, false)

	if err != nil {
		log.Println(err)
		return err
	}

	err = ePlugin.Start()
	if err != nil {
		log.Println(err)
		return err
	}

	var resp *plugin.Response
	resp, err = plugin.WaitForResponse(ePlugin, time.Second*3)

	if err != nil {
		log.Println(err)
		return err
	}

	if resp.State != plugin.PluginSuccess {
		log.Println("Plugin loading did not succeed: %s\n", resp.ErrorMessage)
		return fmt.Errorf("Plugin loading did not succeed: %s\n", resp.ErrorMessage)
	}

	lPlugin.Meta = resp.Meta
	lPlugin.Type = resp.Type
	lPlugin.Token = resp.Token
	lPlugin.LoadedTime = time.Now()
	lPlugin.State = LoadedState

	p.LoadedPlugins = append(p.LoadedPlugins, lPlugin)

	return nil
}
