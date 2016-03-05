package cfgfile

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	. "github.com/smartystreets/goconvey/convey"
)

type testConfig struct {
	Foo string
	Bar string
}

func writeYaml(tc testConfig) string {
	b, err := yaml.Marshal(tc)
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile("", "yaml.conf")
	if err != nil {
		panic(err)
	}
	if _, err := f.Write(b); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fp, err := filepath.Abs(f.Name())
	if err != nil {
		panic(err)
	}
	return fp
}

func writeJson(tc testConfig) string {
	b, err := json.Marshal(tc)
	if err != nil {
		panic(err)
	}
	f, err := ioutil.TempFile("", "json.conf")
	if err != nil {
		panic(err)
	}
	if _, err := f.Write(b); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fp, err := filepath.Abs(f.Name())
	if err != nil {
		panic(err)
	}
	return fp
}

func TestReadConfig(t *testing.T) {
	tc := testConfig{"Tom", "Justin"}
	yamlFile := writeYaml(tc)
	jsonFile := writeJson(tc)

	Convey("Unmarshal yaml file", t, func() {
		config := testConfig{}
		err := Read(yamlFile, &config)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})

	Convey("Unmarshal json file", t, func() {
		config := testConfig{}
		err := Read(jsonFile, &config)
		So(err, ShouldBeNil)
		So(config.Foo, ShouldResemble, "Tom")
		So(config.Bar, ShouldResemble, "Justin")
	})
}
