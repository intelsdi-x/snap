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
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/pkg/fileutils"
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

func CheckPluginType(id PluginType) bool {
	pts := map[PluginType]string{
		0: "collector",
		1: "processor",
		2: "publisher",
	}

	_, ok := pts[id]

	return ok
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
	Port() string
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
	path      string
	checkSum  [sha256.Size]byte
	signature []byte
}

// NewRequestedPlugin returns a Requested Plugin which represents the plugin path and signature
// It takes the full path of the plugin (path), temp path (fileName), and content of the file (b) and returns a requested plugin and error
// The argument b (content of the file) can be nil
func NewRequestedPlugin(path, fileName string, b []byte) (*RequestedPlugin, error) {
	var rp *RequestedPlugin
	// this case is for the snaptel cli as b is unknown and needs to be read
	if b == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return nil, err
		}
		size := info.Size()
		bytes := make([]byte, size)
		buffer := bufio.NewReader(file)
		_, err = buffer.Read(bytes)
		if err != nil {
			return nil, err
		}
		tempFile, err := fileutils.WriteFile(filepath.Base(path), fileName, bytes)
		if err != nil {
			return nil, err
		}
		rp = &RequestedPlugin{
			path:      tempFile,
			signature: nil,
		}
		genErr := rp.generateCheckSum()
		if genErr != nil {
			return nil, err
		}
	} else {
		// this is for the REST API as the content is read earlier
		tmpFile, err := fileutils.WriteFile(filepath.Base(path), fileName, b)
		if err != nil {
			return nil, nil
		}
		rp = &RequestedPlugin{
			path:      tmpFile,
			signature: nil,
		}
		genErr := rp.generateCheckSum()
		if genErr != nil {
			return nil, err
		}
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

func (p *RequestedPlugin) SetPath(path string) {
	p.path = path
}

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

func (p *RequestedPlugin) ReadSignatureFile(file string) error {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	p.SetSignature(b)
	return nil
}
