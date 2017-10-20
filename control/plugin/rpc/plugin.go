package rpc

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

var (
	rpcLogger = log.WithFields(log.Fields{
		"_module": "rpc",
	})
)

// NewGetConfigPolicyReply given a config *cpolicy.ConfigPolicy returns a GetConfigPolicyReply.
func NewGetConfigPolicyReply(policy *cpolicy.ConfigPolicy) (*GetConfigPolicyReply, error) {
	ret := &GetConfigPolicyReply{
		BoolPolicy:    map[string]*BoolPolicy{},
		FloatPolicy:   map[string]*FloatPolicy{},
		IntegerPolicy: map[string]*IntegerPolicy{},
		StringPolicy:  map[string]*StringPolicy{},
	}

	for _, node := range policy.GetAll() {
		key := strings.Join(node.Key, ".")

		for _, rule := range node.RulesAsTable() {
			switch rule.Type {
			case cpolicy.BoolType:
				r := &BoolRule{
					Required: rule.Required,
				}
				if rule.Default != nil {
					r.Default = rule.Default.(ctypes.ConfigValueBool).Value
				}
				if ret.BoolPolicy[key] == nil {
					ret.BoolPolicy[key] = &BoolPolicy{
						Rules: map[string]*BoolRule{},
						Key:   node.Key,
					}
				}
				ret.BoolPolicy[key].Rules[rule.Name] = r
			case cpolicy.StringType:
				r := &StringRule{
					Required: rule.Required,
				}
				if rule.Default != nil {
					r.Default = rule.Default.(ctypes.ConfigValueStr).Value
				}
				if ret.StringPolicy[key] == nil {
					ret.StringPolicy[key] = &StringPolicy{
						Rules: map[string]*StringRule{},
						Key:   node.Key,
					}
				}
				ret.StringPolicy[key].Rules[rule.Name] = r
			case cpolicy.IntegerType:
				r := &IntegerRule{
					Required: rule.Required,
				}
				if rule.Default != nil {
					r.Default = int64(rule.Default.(ctypes.ConfigValueInt).Value)
				}
				if rule.Maximum != nil {
					r.Maximum = int64(rule.Maximum.(ctypes.ConfigValueInt).Value)
				}
				if rule.Minimum != nil {
					r.Minimum = int64(rule.Minimum.(ctypes.ConfigValueInt).Value)
				}
				if ret.IntegerPolicy[key] == nil {
					ret.IntegerPolicy[key] = &IntegerPolicy{
						Rules: map[string]*IntegerRule{},
						Key:   node.Key,
					}
				}
				ret.IntegerPolicy[key].Rules[rule.Name] = r
			case cpolicy.FloatType:
				r := &FloatRule{
					Required: rule.Required,
				}
				if rule.Default != nil {
					r.Default = rule.Default.(ctypes.ConfigValueFloat).Value
				}
				if rule.Maximum != nil {
					r.Maximum = rule.Maximum.(ctypes.ConfigValueFloat).Value
				}
				if rule.Minimum != nil {
					r.Minimum = rule.Minimum.(ctypes.ConfigValueFloat).Value
				}
				if ret.FloatPolicy[key] == nil {
					ret.FloatPolicy[key] = &FloatPolicy{
						Rules: map[string]*FloatRule{},
						Key:   node.Key,
					}
				}
				ret.FloatPolicy[key].Rules[rule.Name] = r
			}

		}
	}
	return ret, nil
}

func ToConfigPolicy(reply *GetConfigPolicyReply) *cpolicy.ConfigPolicy {
	result := cpolicy.New()
	nodes := make(map[string]*cpolicy.ConfigPolicyNode)
	for k, v := range reply.BoolPolicy {
		if _, ok := nodes[k]; !ok {
			nodes[k] = cpolicy.NewPolicyNode()
		}
		for key, val := range v.Rules {
			var br *cpolicy.BoolRule
			var err error
			if val.HasDefault {
				br, err = cpolicy.NewBoolRule(key, val.Required, val.Default)
			} else {
				br, err = cpolicy.NewBoolRule(key, val.Required)
			}
			if err != nil {
				rpcLogger.Warn("Empty key found with value %v", val)
				continue
			}
			nodes[k].Add(br)
		}
	}

	for k, v := range reply.StringPolicy {
		if _, ok := nodes[k]; !ok {
			nodes[k] = cpolicy.NewPolicyNode()
		}
		for key, val := range v.Rules {
			var sr *cpolicy.StringRule
			var err error
			if val.HasDefault {
				sr, err = cpolicy.NewStringRule(key, val.Required, val.Default)
			} else {
				sr, err = cpolicy.NewStringRule(key, val.Required)
			}
			if err != nil {
				rpcLogger.Warn("Empty key found with value %v", val)
				continue
			}

			nodes[k].Add(sr)
		}
	}

	for k, v := range reply.IntegerPolicy {
		if _, ok := nodes[k]; !ok {
			nodes[k] = cpolicy.NewPolicyNode()
		}
		for key, val := range v.Rules {
			var ir *cpolicy.IntRule
			var err error
			if val.HasDefault {
				ir, err = cpolicy.NewIntegerRule(key, val.Required, int(val.Default))
			} else {
				ir, err = cpolicy.NewIntegerRule(key, val.Required)
			}
			if err != nil {
				rpcLogger.Warn("Empty key found with value %v", val)
				continue
			}
			if val.HasMin {
				ir.SetMinimum(int(val.Minimum))
			}
			if val.HasMax {
				ir.SetMaximum(int(val.Maximum))
			}

			nodes[k].Add(ir)
		}
	}

	for k, v := range reply.FloatPolicy {
		if _, ok := nodes[k]; !ok {
			nodes[k] = cpolicy.NewPolicyNode()
		}
		for key, val := range v.Rules {
			var fr *cpolicy.FloatRule
			var err error
			if val.HasDefault {
				fr, err = cpolicy.NewFloatRule(key, val.Required, val.Default)
			} else {

				fr, err = cpolicy.NewFloatRule(key, val.Required)
			}
			if err != nil {
				rpcLogger.Warn("Empty key found with value %v", val)
				continue
			}

			if val.HasMin {
				fr.SetMinimum(val.Minimum)
			}
			if val.HasMax {
				fr.SetMaximum(val.Maximum)
			}

			nodes[k].Add(fr)
		}
	}

	for key, node := range nodes {
		var keys []string
		// if the []string is present, use it.
		// if not, fall back to dot separated key
		if val, ok := reply.BoolPolicy[key]; ok && val != nil && val.Key != nil {
			keys = val.Key
		} else if val, ok := reply.StringPolicy[key]; ok && val != nil && val.Key != nil {
			keys = val.Key
		} else if val, ok := reply.FloatPolicy[key]; ok && val != nil && val.Key != nil {
			keys = val.Key
		} else if val, ok := reply.IntegerPolicy[key]; ok && val != nil && val.Key != nil {
			keys = val.Key
		} else {
			keys = strings.Split(key, ".")
		}
		result.Add(keys, node)
	}

	return result
}
