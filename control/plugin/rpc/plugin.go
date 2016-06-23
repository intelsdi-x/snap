package rpc

import (
	"strings"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

// NewGetConfigPolicyReply given a config *cpolicy.ConfigPolicy returns a GetConfigPolicyReply.
func NewGetConfigPolicyReply(policy *cpolicy.ConfigPolicy) (*GetConfigPolicyReply, error) {
	ret := &GetConfigPolicyReply{
		BoolPolicy:    map[string]*BoolPolicy{},
		FloatPolicy:   map[string]*FloatPolicy{},
		IntegerPolicy: map[string]*IntegerPolicy{},
		StringPolicy:  map[string]*StringPolicy{},
	}

	for key, node := range policy.GetAll() {
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
					ret.BoolPolicy[key] = &BoolPolicy{Rules: map[string]*BoolRule{}}
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
					ret.StringPolicy[key] = &StringPolicy{Rules: map[string]*StringRule{}}
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
					ret.IntegerPolicy[key] = &IntegerPolicy{Rules: map[string]*IntegerRule{}}
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
					ret.FloatPolicy[key] = &FloatPolicy{Rules: map[string]*FloatRule{}}
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
			br, err := cpolicy.NewBoolRule(key, val.Required, val.Default)
			if err != nil {
				// The only error that can be thrown is empty key error, ignore something with empty key
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
			sr, err := cpolicy.NewStringRule(key, val.Required, val.Default)
			if err != nil {
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
			sr, err := cpolicy.NewIntegerRule(key, val.Required, int(val.Default))
			if err != nil {
				continue
			}

			nodes[k].Add(sr)
		}
	}

	for k, v := range reply.FloatPolicy {
		if _, ok := nodes[k]; !ok {
			nodes[k] = cpolicy.NewPolicyNode()
		}
		for key, val := range v.Rules {
			sr, err := cpolicy.NewFloatRule(key, val.Required, val.Default)
			if err != nil {
				continue
			}

			nodes[k].Add(sr)
		}
	}

	for key, node := range nodes {
		keys := strings.Split(key, ".")
		result.Add(keys, node)
	}

	return result
}
