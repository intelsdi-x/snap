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

// Code generated by protoc-gen-go.
// source: github.com/intelsdi-x/snap/internal/common/common.proto
// DO NOT EDIT!

/*
Package common is a generated protocol buffer package.

It is generated from these files:
	github.com/intelsdi-x/snap/internal/common/common.proto

It has these top-level messages:
	Time
	Empty
	ConfigMap
*/
package common

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type Time struct {
	Sec  int64 `protobuf:"varint,1,opt,name=sec" json:"sec,omitempty"`
	Nsec int64 `protobuf:"varint,2,opt,name=nsec" json:"nsec,omitempty"`
}

func (m *Time) Reset()                    { *m = Time{} }
func (m *Time) String() string            { return proto.CompactTextString(m) }
func (*Time) ProtoMessage()               {}
func (*Time) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type ConfigMap struct {
	IntMap    map[string]int64   `protobuf:"bytes,1,rep,name=IntMap" json:"IntMap,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	StringMap map[string]string  `protobuf:"bytes,2,rep,name=StringMap" json:"StringMap,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	FloatMap  map[string]float64 `protobuf:"bytes,3,rep,name=FloatMap" json:"FloatMap,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"fixed64,2,opt,name=value"`
	BoolMap   map[string]bool    `protobuf:"bytes,4,rep,name=BoolMap" json:"BoolMap,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *ConfigMap) Reset()                    { *m = ConfigMap{} }
func (m *ConfigMap) String() string            { return proto.CompactTextString(m) }
func (*ConfigMap) ProtoMessage()               {}
func (*ConfigMap) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ConfigMap) GetIntMap() map[string]int64 {
	if m != nil {
		return m.IntMap
	}
	return nil
}

func (m *ConfigMap) GetStringMap() map[string]string {
	if m != nil {
		return m.StringMap
	}
	return nil
}

func (m *ConfigMap) GetFloatMap() map[string]float64 {
	if m != nil {
		return m.FloatMap
	}
	return nil
}

func (m *ConfigMap) GetBoolMap() map[string]bool {
	if m != nil {
		return m.BoolMap
	}
	return nil
}

func init() {
	proto.RegisterType((*Time)(nil), "common.Time")
	proto.RegisterType((*Empty)(nil), "common.Empty")
	proto.RegisterType((*ConfigMap)(nil), "common.ConfigMap")
}

var fileDescriptor0 = []byte{
	// 307 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x92, 0xcd, 0x4a, 0xc3, 0x40,
	0x14, 0x85, 0x49, 0xd3, 0xe6, 0xe7, 0x56, 0x45, 0x06, 0x17, 0x52, 0x50, 0x4b, 0x56, 0x5d, 0x68,
	0x02, 0x8a, 0x58, 0x5b, 0x71, 0xa1, 0x54, 0x70, 0xe1, 0x26, 0xfa, 0x02, 0x49, 0x4d, 0xeb, 0xe0,
	0x64, 0x26, 0x24, 0x53, 0x31, 0xcf, 0xec, 0x4b, 0x98, 0xf9, 0x69, 0x99, 0x92, 0x45, 0x56, 0x39,
	0x67, 0xee, 0xf9, 0xb8, 0x9c, 0x4b, 0xe0, 0x6e, 0x8d, 0xf9, 0xd7, 0x26, 0x0d, 0x97, 0x2c, 0x8f,
	0x30, 0xe5, 0x19, 0xa9, 0x3e, 0xf1, 0xd5, 0x6f, 0x54, 0xd1, 0xa4, 0x90, 0xbe, 0xa4, 0x09, 0x89,
	0x9a, 0x61, 0xce, 0xa8, 0xfe, 0x84, 0x45, 0xc9, 0x38, 0x43, 0x8e, 0x72, 0xc1, 0x25, 0xf4, 0x3f,
	0x70, 0x9e, 0xa1, 0x63, 0xb0, 0xab, 0x6c, 0x79, 0x6a, 0x8d, 0xad, 0x89, 0x1d, 0x0b, 0x89, 0x10,
	0xf4, 0xa9, 0x78, 0xea, 0xc9, 0x27, 0xa9, 0x03, 0x17, 0x06, 0x8b, 0xbc, 0xe0, 0x75, 0xf0, 0x67,
	0x83, 0xff, 0xcc, 0xe8, 0x0a, 0xaf, 0xdf, 0x92, 0x02, 0xdd, 0x82, 0xf3, 0x4a, 0x79, 0xa3, 0x1a,
	0xde, 0x9e, 0x0c, 0xaf, 0xcf, 0x42, 0xbd, 0x6b, 0x17, 0x09, 0xd5, 0x7c, 0x41, 0x79, 0x59, 0xc7,
	0x0e, 0x96, 0x06, 0x3d, 0x82, 0xff, 0xce, 0x4b, 0x4c, 0x45, 0xa0, 0x59, 0x23, 0xc8, 0x71, 0x9b,
	0xdc, 0x45, 0x14, 0xec, 0x57, 0x5b, 0x8f, 0xe6, 0xe0, 0xbd, 0x10, 0x96, 0xc8, 0xc5, 0xb6, 0xc4,
	0x2f, 0xda, 0xf8, 0x36, 0xa1, 0x68, 0x6f, 0xa5, 0x2d, 0x9a, 0x82, 0xfb, 0xc4, 0x18, 0x11, 0x6c,
	0x5f, 0xb2, 0xe7, 0x6d, 0x56, 0x07, 0x14, 0xea, 0xa6, 0xca, 0x8d, 0xee, 0x61, 0x68, 0xb4, 0x11,
	0x97, 0xfb, 0xce, 0x6a, 0x79, 0x39, 0x3f, 0x16, 0x12, 0x9d, 0xc0, 0xe0, 0x27, 0x21, 0x9b, 0x4c,
	0x9f, 0x4e, 0x99, 0x59, 0x6f, 0x6a, 0x8d, 0x1e, 0xe0, 0x68, 0xbf, 0x4e, 0x17, 0xed, 0x9b, 0xf4,
	0x1c, 0x0e, 0xf7, 0xda, 0x74, 0xc1, 0x96, 0x09, 0xcf, 0xe0, 0xc0, 0xac, 0xd3, 0xc5, 0x7a, 0x06,
	0x9b, 0x3a, 0xf2, 0x9f, 0xb9, 0xf9, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x5e, 0xea, 0x16, 0xdb, 0x6e,
	0x02, 0x00, 0x00,
}