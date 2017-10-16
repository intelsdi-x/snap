/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package rpc

import (
	"errors"
	"strconv"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/grpc/common"
	log "github.com/sirupsen/logrus"
)

type Metric struct {
	Namespace          []string                  `json:"namespace"`
	Version            int                       `json:"version"`
	LastAdvertisedTime time.Time                 `json:"last_advertised_time"`
	ConfigPolicy       *cpolicy.ConfigPolicyNode `json:"policy"`
}

type LoadedPlugin struct {
	Name            string                `json:"name"`
	Version         int                   `json:"version"`
	TypeName        string                `json:"type_name"`
	IsSigned        bool                  `json:"signed"`
	Status          string                `json:"status"`
	LoadedTimestamp time.Time             `json:"timestamp"`
	ConfigPolicy    *cpolicy.ConfigPolicy `json:"policy,omitempty"`
}

type AvailablePlugin struct {
	Name     string    `json:"name"`
	Version  int       `json:"version"`
	TypeName string    `json:"type_name"`
	Signed   bool      `json:"signed"`
	HitCount int       `json:"hit_count"`
	ID       uint32    `json:"id"`
	LastHit  time.Time `json:"last_hit"`
}

func ConvertSnapErrors(s []*common.SnapError) []serror.SnapError {
	rerrs := make([]serror.SnapError, len(s))
	for i, err := range s {
		rerrs[i] = serror.New(errors.New(err.ErrorString), common.GetFields(err))
	}
	return rerrs
}

func NewErrors(errs []serror.SnapError) []*common.SnapError {
	errors := make([]*common.SnapError, len(errs))
	for i, err := range errs {
		fields := make(map[string]string)
		for k, v := range err.Fields() {
			switch t := v.(type) {
			case string:
				fields[k] = t
			case int:
				fields[k] = strconv.Itoa(t)
			case float64:
				fields[k] = strconv.FormatFloat(t, 'f', -1, 64)
			default:
				log.Errorf("Unexpected type %v\n", t)
			}
		}
		errors[i] = &common.SnapError{ErrorFields: fields, ErrorString: err.Error()}
	}
	return errors
}
