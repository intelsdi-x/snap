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

package core

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/intelsdi-x/snap/core/cdata"
)

// Plugin interface describing a plugin type, name and version
type Plugin interface {
	TypeName() string
	Name() string
	Version() int
}

// PluginType is used as an enum to differentiate types of snap plugins
type PluginType int

// ToPluginType transforms the name, e.g. "collector", to the corresponding enum
func ToPluginType(name string) (PluginType, error) {
	pts := map[string]PluginType{
		"collector": 0,
		"processor": 1,
		"publisher": 2,
	}
	t, ok := pts[name]
	if !ok {
		return -1, fmt.Errorf("invalid plugin type name given %s", name)
	}
	return t, nil
}

// String returns a string representation of collector, processor
// and publisher of a plugin type.
func (pt PluginType) String() string {
	return []string{
		"collector",
		"processor",
		"publisher",
	}[pt]
}

// enum of plugin types
const (
	CollectorPluginType PluginType = iota
	ProcessorPluginType
	PublisherPluginType
)

// AvailablePlugin is the public interface for a running plugin
type AvailablePlugin interface {
	Plugin
	HitCount() int
	LastHit() time.Time
	ID() uint32
}

// CatalogedPlugin the public contract for a plugin
type CatalogedPlugin interface {
	Plugin
	IsSigned() bool
	Status() string
	PluginPath() string
	LoadedTimestamp() *time.Time
}

// PluginCatalog is the collection of _loaded_ plugins
type PluginCatalog []CatalogedPlugin

// SubscribedPlugin is used when ./scheduler calls SubscribeDeps on control.
type SubscribedPlugin interface {
	Plugin
	Config() *cdata.ConfigDataNode
}

// RequestedPlugin is used by the API when loading a plugin.
type RequestedPlugin struct {
	path      string
	checkSum  [sha256.Size]byte
	signature []byte
}

// NewRequestedPlugin creates a requested plugin
func NewRequestedPlugin(path string) (*RequestedPlugin, error) {
	rp := &RequestedPlugin{
		path:      path,
		signature: nil,
	}
	err := rp.generateCheckSum()
	if err != nil {
		return nil, err
	}
	return rp, nil
}

// Path returns the plugin request path
func (p *RequestedPlugin) Path() string {
	return p.path
}

// CheckSum returns the plugin crequest checksum
func (p *RequestedPlugin) CheckSum() [sha256.Size]byte {
	return p.checkSum
}

// Signature returns the plugin singnature
func (p *RequestedPlugin) Signature() []byte {
	return p.signature
}

// SetPath sets the plugin path
func (p *RequestedPlugin) SetPath(path string) {
	p.path = path
}

// SetSignature sets plugin signature
func (p *RequestedPlugin) SetSignature(data []byte) {
	p.signature = data
}

func (p *RequestedPlugin) generateCheckSum() error {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(p.path); err != nil {
		return err
	}
	p.checkSum = sha256.Sum256(b)
	return nil
}

// ReadSignatureFile reads signature file and sets signature
func (p *RequestedPlugin) ReadSignatureFile(file string) error {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	p.SetSignature(b)
	return nil
}
