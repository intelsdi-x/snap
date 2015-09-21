package opentsdb

import (
	"bytes"
	"encoding/json"
	"testing"
)

var stringtests = []struct {
	sv   StringValue
	json []byte
}{
	{StringValue("dog-cat-22"), []byte(`"dog-cat-22"`)},
	{StringValue("dog_cat_22"), []byte(`"dog__cat__22"`)},
	{StringValue("http://google.com:8080"), []byte(`"http_.//google.com_.8080"`)},
	{StringValue("日"), []byte(`"_E6_97_A5"`)},
	{StringValue("/psutil/load/load15"), []byte(`"/psutil/load/load15"`)},
	{StringValue("/psutil/vm/free"), []byte(`"/psutil/vm/free"`)},
}

func TestStringValueMarshaling(t *testing.T) {
	for i, tt := range stringtests {
		json, err := json.Marshal(tt.sv)
		if err != nil {
			t.Errorf("%d. Marshal(%q) returned err: %s", i, tt.sv, err)
		} else {
			if !bytes.Equal(json, tt.json) {
				t.Errorf(
					"%d. Marshal(%q) => %q, want %q",
					i, tt.sv, json, tt.json,
				)
			}
		}
	}
}

func TestStringValueUnMarshaling(t *testing.T) {
	for i, tt := range stringtests {
		var sv StringValue
		err := json.Unmarshal(tt.json, &sv)
		if err != nil {
			t.Errorf("%d. Unmarshal(%q, &str) returned err: %s", i, tt.json, err)
		} else {
			if sv != tt.sv {
				t.Errorf(
					"%d. Unmarshal(%q, &str) => str==%q, want %q",
					i, tt.json, sv, tt.sv,
				)
			}
		}
	}
}
