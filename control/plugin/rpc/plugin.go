package rpc

import (
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
					r.Default = rule.Default.(*ctypes.ConfigValueBool).Value
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
					r.Default = rule.Default.(*ctypes.ConfigValueStr).Value
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
					r.Default = int64(rule.Default.(*ctypes.ConfigValueInt).Value)
				}
				if rule.Maximum != nil {
					r.Maximum = int64(rule.Maximum.(*ctypes.ConfigValueInt).Value)
				}
				if rule.Minimum != nil {
					r.Minimum = int64(rule.Minimum.(*ctypes.ConfigValueInt).Value)
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
					r.Default = rule.Default.(*ctypes.ConfigValueFloat).Value
				}
				if rule.Maximum != nil {
					r.Maximum = rule.Maximum.(*ctypes.ConfigValueFloat).Value
				}
				if rule.Minimum != nil {
					r.Minimum = rule.Minimum.(*ctypes.ConfigValueFloat).Value
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
