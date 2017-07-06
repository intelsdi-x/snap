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
	"strconv"
	"strings"
	"time"

	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/serror"
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
		"collector":           0,
		"processor":           1,
		"publisher":           2,
		"streaming-collector": 3,
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
		3: "streaming-collector",
	}

	_, ok := pts[id]

	return ok
}

func GetPluginType(t string) (PluginType, error) {
	if ityp, err := strconv.Atoi(t); err == nil {
		if !CheckPluginType(PluginType(ityp)) {
			return PluginType(-1), fmt.Errorf("invalid plugin type id given %d", ityp)
		}
		return PluginType(ityp), nil
	}
	return ToPluginType(t)
}

func (pt PluginType) String() string {
	return []string{
		"collector",
		"processor",
		"publisher",
		"streaming-collector",
	}[pt]
}

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	ProcessorPluginType
	PublisherPluginType
	StreamingCollectorPluginType
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
	Key() string
}

// the collection of cataloged plugins used
// by mgmt modules
type PluginCatalog []CatalogedPlugin

type SubscribedPlugin interface {
	Plugin
	Config() *cdata.ConfigDataNode
}

type SubscribedPluginAssert func(plugins []SubscribedPlugin) serror.SnapError

type RequestedPlugin struct {
	path        string
	checkSum    [sha256.Size]byte
	signature   []byte
	certPath    string
	keyPath     string
	caCertPaths string
	tlsEnabled  bool
	autoLoaded  bool
	uri         *url.URL
}

// NewRequestedPlugin returns a Requested Plugin which represents the plugin path and signature
// It takes the full path of the plugin (path), temp path (fileName), and content of the file (b) and returns a requested plugin and error
// The argument b (content of the file) can be nil
func NewRequestedPlugin(path, fileName string, b []byte) (*RequestedPlugin, error) {
	// Checks if string is URL
	if IsUri(path) {
		if uri, err := url.ParseRequestURI(path); err == nil && uri != nil {
			return &RequestedPlugin{uri: uri}, nil
		}
	}
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

// Checks if string is URL
func IsUri(url string) bool {
	if !govalidator.IsURL(url) || !strings.HasPrefix(url, "http") {
		return false
	}
	return true
}

func (p *RequestedPlugin) Path() string {
	return p.path
}

// CertPath returns the path to TLS certificate for requested plugin to use
func (p *RequestedPlugin) CertPath() string {
	return p.certPath
}

// KeyPath returns the path to TLS key for requested plugin to use
func (p *RequestedPlugin) KeyPath() string {
	return p.keyPath
}

// CACertPaths returns the list of TLS CA cert paths for plugin to use
func (p *RequestedPlugin) CACertPaths() string {
	return p.caCertPaths
}

// TLSEnabled returns the TLS enabled flag for requested plugin
func (p *RequestedPlugin) TLSEnabled() bool {
	return p.tlsEnabled
}

func (p *RequestedPlugin) CheckSum() [sha256.Size]byte {
	return p.checkSum
}

func (p *RequestedPlugin) Signature() []byte {
	return p.signature
}

func (p *RequestedPlugin) Uri() *url.URL {
	return p.uri
}

func (p *RequestedPlugin) SetPath(path string) {
	p.path = path
}

// SetCertPath sets the path to TLS certificate for requested plugin to use
func (p *RequestedPlugin) SetCertPath(certPath string) {
	p.certPath = certPath
}

// SetKeyPath sets the path to TLS key for requested plugin to use
func (p *RequestedPlugin) SetKeyPath(keyPath string) {
	p.keyPath = keyPath
}

// SetCACertPaths sets the list of paths to TLS CA certificate for plugin to use
func (p *RequestedPlugin) SetCACertPaths(caCertPaths string) {
	p.caCertPaths = caCertPaths
}

// SetTLSEnabled sets the TLS flag on requested plugin
func (p *RequestedPlugin) SetTLSEnabled(tlsEnabled bool) {
	p.tlsEnabled = tlsEnabled
}

func (p *RequestedPlugin) SetSignature(data []byte) {
	p.signature = data
}

func (p *RequestedPlugin) SetUri(uri *url.URL) {
	p.uri = uri
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
