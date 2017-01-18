package rest

import (
	"encoding/json"
	"fmt"
)

// default configuration values
const (
	defaultEnable          bool   = true
	defaultPort            int    = 8181
	defaultAddress         string = ""
	defaultHTTPS           bool   = false
	defaultRestCertificate string = ""
	defaultRestKey         string = ""
	defaultAuth            bool   = false
	defaultAuthPassword    string = ""
	defaultPortSetByConfig bool   = false
	defaultPprof           bool   = false
)

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	Enable           bool   `json:"enable"yaml:"enable"`
	Port             int    `json:"port"yaml:"port"`
	Address          string `json:"addr"yaml:"addr"`
	HTTPS            bool   `json:"https"yaml:"https"`
	RestCertificate  string `json:"rest_certificate"yaml:"rest_certificate"`
	RestKey          string `json:"rest_key"yaml:"rest_key"`
	RestAuth         bool   `json:"rest_auth"yaml:"rest_auth"`
	RestAuthPassword string `json:"rest_auth_password"yaml:"rest_auth_password"`
	portSetByConfig  bool   ``
	Pprof            bool   `json:"pprof"yaml:"pprof"`
}

const (
	CONFIG_CONSTRAINTS = `
			"restapi" : {
				"type": ["object", "null"],
				"properties" : {
					"enable": {
						"type": "boolean"
					},
					"https" : {
						"type": "boolean"
					},
					"rest_auth": {
						"type": "boolean"
					},
					"rest_auth_password": {
						"type": "string"
					},
					"rest_certificate": {
						"type": "string"
					},
					"rest_key" : {
						"type": "string"
					},
					"port" : {
						"type": "integer",
						"minimum": 1,
						"maximum": 65535
					},
					"addr" : {
						"type": "string"
					},
					"pprof": {
						"type": "boolean"
					}
				},
				"additionalProperties": false
			}
	`
)

// GetDefaultConfig gets the default snapteld configuration
func GetDefaultConfig() *Config {
	return &Config{
		Enable:           defaultEnable,
		Port:             defaultPort,
		Address:          defaultAddress,
		HTTPS:            defaultHTTPS,
		RestCertificate:  defaultRestCertificate,
		RestKey:          defaultRestKey,
		RestAuth:         defaultAuth,
		RestAuthPassword: defaultAuthPassword,
		portSetByConfig:  defaultPortSetByConfig,
		Pprof:            defaultPprof,
	}
}

// define a method that can be used to determine if the port the RESTful
// API is listening on was set in the configuration file
func (c *Config) PortSetByConfigFile() bool {
	return c.portSetByConfig
}

// UnmarshalJSON unmarshals valid json into a Config.  An example Config can be found
// at github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (c *Config) UnmarshalJSON(data []byte) error {
	// construct a map of strings to json.RawMessages (to defer the parsing of individual
	// fields from the unmarshalled interface until later) and unmarshal the input
	// byte array into that map
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// loop through the individual map elements, parse each in turn, and set
	// the appropriate field in this configuration
	for k, v := range t {
		switch k {
		case "enable":
			if err := json.Unmarshal(v, &(c.Enable)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::enable')", err)
			}
		case "port":
			if err := json.Unmarshal(v, &(c.Port)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::port')", err)
			}
			c.portSetByConfig = true
		case "addr":
			if err := json.Unmarshal(v, &(c.Address)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::addr')", err)
			}
		case "https":
			if err := json.Unmarshal(v, &(c.HTTPS)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::https')", err)
			}
		case "rest_certificate":
			if err := json.Unmarshal(v, &(c.RestCertificate)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_certificate')", err)
			}
		case "rest_key":
			if err := json.Unmarshal(v, &(c.RestKey)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_key')", err)
			}
		case "rest_auth":
			if err := json.Unmarshal(v, &(c.RestAuth)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_auth')", err)
			}
		case "rest_auth_password":
			if err := json.Unmarshal(v, &(c.RestAuthPassword)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_auth_password')", err)
			}
		case "pprof":
			if err := json.Unmarshal(v, &(c.Pprof)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::pprof')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file while parsing 'restapi'", k)
		}
	}
	return nil
}
