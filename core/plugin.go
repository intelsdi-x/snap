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

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/cdata"
)

type Plugin interface {
	TypeName() string
	Name() string
	Version() int
}

type PluginType int

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

func (pt PluginType) String() string {
	return []string{
		"collector",
		"processor",
		"publisher",
	}[pt]
}

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	ProcessorPluginType
	PublisherPluginType
)

type AvailablePlugin interface {
	Plugin
	HitCount() int
	LastHit() time.Time
	ID() uint32
}

// the public interface for a plugin
// this should be the contract for
// how mgmt modules know a plugin
type CatalogedPlugin interface {
	Plugin
	IsSigned() bool
	Status() string
	PluginPath() string
	LoadedTimestamp() *time.Time
	Policy() *cpolicy.ConfigPolicy
}

// the collection of cataloged plugins used
// by mgmt modules
type PluginCatalog []CatalogedPlugin

type SubscribedPlugin interface {
	Plugin
	Config() *cdata.ConfigDataNode
}

type RequestedPlugin struct {
	path       string
	checkSum   [sha256.Size]byte
	signature  []byte
	autoLoaded bool
}

func NewRequestedPlugin(path string) (*RequestedPlugin, error) {
	rp := &RequestedPlugin{
		path:       path,
		signature:  nil,
		autoLoaded: true,
	}
	err := rp.generateCheckSum()
	if err != nil {
		return nil, err
	}
	return rp, nil
}

func (p *RequestedPlugin) Path() string {
	return p.path
}

func (p *RequestedPlugin) CheckSum() [sha256.Size]byte {
	return p.checkSum
}

func (p *RequestedPlugin) Signature() []byte {
	return p.signature
}

func (p *RequestedPlugin) AutoLoaded() bool {
	return p.autoLoaded
}

func (p *RequestedPlugin) SetPath(path string) {
	p.path = path
}

func (p *RequestedPlugin) SetSignature(data []byte) {
	p.signature = data
}

func (p *RequestedPlugin) SetAutoLoaded(isAutoLoaded bool) {
	p.autoLoaded = isAutoLoaded
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

func (p *RequestedPlugin) ReadSignatureFile(file string) error {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	p.SetSignature(b)
	return nil
}
