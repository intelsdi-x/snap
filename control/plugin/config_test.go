package plugin

import (
	"errors"
	"reflect"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicy(t *testing.T) {
	Convey(".NewConfigPolicy()", t, func() {
		c := NewConfigPolicy()
		Convey("it returns a config policy", func() {
			So(c, ShouldHaveSameTypeAs, &ConfigPolicy{})
		})
	})
	Convey(".Add()", t, func() {
		c := NewConfigPolicy()
		Convey("It adds the policy", func() {
			p1 := &Policy{
				Type:     reflect.String,
				Key:      "test",
				Required: false,
			}
			c.Add("/foo/bar", "/foo/bar/baz", p1)
			So((*c)["/foo/bar/baz"][0], ShouldEqual, p1)
		})
		Convey("it panics if any required policy fields are missing", func() {
			p1 := &Policy{
				Key:      "test",
				Required: false,
			}
			So(func() {
				c.Add("/foo/bar", "bad", p1)
			}, ShouldPanicWith, "Type and Key are required fields on a policy")
		})
		Convey("it panics if an invalid namespace format is given", func() {
			p1 := &Policy{
				Type:     reflect.String,
				Key:      "test",
				Required: false,
			}
			So(func() {
				c.Add("/foo/bar", "bad", p1)
			}, ShouldPanicWith, "config policy namespace must begin with [/]")
		})
		Convey("it panics if the given namespace is not a child to the plugin name", func() {
			p1 := &Policy{
				Type:     reflect.String,
				Key:      "test",
				Required: false,
			}
			So(func() {
				c.Add("/foo/bar", "/not/mine", p1)
			}, ShouldPanicWith, "metric namespace must fall under plugin's namespace")
		})
		Convey("it appends to the collection if a key already has a policy/policies", func() {
			p1 := &Policy{
				Type:     reflect.String,
				Key:      "test",
				Required: false,
			}
			p2 := &Policy{
				Type:     reflect.String,
				Key:      "test2",
				Required: false,
			}
			c.Add("/foo/bar", "/foo/bar/baz", p1)
			c.Add("/foo/bar", "/foo/bar/baz", p2)
			So((*c)["/foo/bar/baz"][1], ShouldEqual, p2)
		})
	})
}

func TestPolicy(t *testing.T) {
	Convey(".Validate()", t, func() {
		Convey("It should return nil if the given type and key are correct", func() {
			p := &Policy{Key: "test", Type: reflect.String}
			pi := &PolicyInput{Key: "test", Value: "hi"}
			err := p.Validate(pi)
			So(err, ShouldBeNil)
		})
		Convey("it should return an error if the wrong key is given", func() {
			p := &Policy{Key: "foo"}
			pi := &PolicyInput{Key: "bar"}
			err := p.Validate(pi)
			So(err, ShouldResemble, errors.New("incorrect key given [bar] for policy [foo]"))
		})
		Convey("it should return an error if the wrong type is given", func() {
			p := &Policy{Key: "foo", Type: reflect.Int}
			pi := &PolicyInput{Key: "foo", Value: "hi"}
			err := p.Validate(pi)
			So(err, ShouldResemble, errors.New("invalid type given for policy foo"))
		})
		Convey("it should return an error if no key is given", func() {
			p := &Policy{Key: "foo"}
			pi := &PolicyInput{}
			err := p.Validate(pi)
			So(err, ShouldResemble, errors.New("incorrect key given [] for policy [foo]"))
		})
		Convey("it should return an error if no value is given", func() {
			p := &Policy{Key: "foo", Type: reflect.Int}
			pi := &PolicyInput{Key: "foo"}
			err := p.Validate(pi)
			So(err, ShouldResemble, errors.New("policy input with nil value given"))
		})
	})
}
