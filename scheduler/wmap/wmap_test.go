package wmap

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflow(t *testing.T) {
	jsonP, _ := ioutil.ReadFile("./sample/1.json")
	yamlP, _ := ioutil.ReadFile("./sample/1.yml")

	Convey("Workflow map", t, func() {

		Convey("from yaml", func() {
			fmt.Println("YAML ---")
			wmap, err := FromYaml(yamlP)
			So(err, ShouldBeNil)
			So(wmap, ShouldNotBeNil)
			fmt.Println(wmap)
		})

		Convey("from json", func() {
			fmt.Println("JSON ---")
			wmap, err := FromJson(jsonP)
			So(err, ShouldBeNil)
			So(wmap, ShouldNotBeNil)
			fmt.Println(wmap)
		})
	})
}
