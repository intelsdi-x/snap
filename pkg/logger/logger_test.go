package logger

import (
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type MockWriter struct {
	Messages string
}

func (m *MockWriter) Write(b []byte) (int, error) {
	m.Messages += string(b)
	return 0, nil
}

// The extra "%s" on Errorf and Fatalf is because go vet hates this one cool trick
// It only looks like this in this test...
func TestLogger(t *testing.T) {
	Convey("Logging", t, func() {
		Convey("defaults to std err", func() {
			logger = nil // need to reset global
			Debug("FOO")
			So(logger, ShouldNotBeNil)
		})

		Convey("sets to output", func() {
			logger = nil // need to reset global
			m := &MockWriter{}
			Output = m
			SetLevel(DebugLevel)
			Debug("FOO")
			So(m.Messages, ShouldContainSubstring, "FOO")
		})

		Convey("debug level", func() {
			m := &MockWriter{}
			Output = m
			SetLevel(DebugLevel)
			logger = log.New(Output, "", log.LstdFlags)
			So(Output, ShouldEqual, m)

			Debug("sheep", "red")
			So(m.Messages, ShouldContainSubstring, "sheep")
			So(m.Messages, ShouldContainSubstring, "red")

			Debugf("cow", "blue")
			So(m.Messages, ShouldContainSubstring, "cow")
			So(m.Messages, ShouldContainSubstring, "blue")

			Info("emu", "green")
			So(m.Messages, ShouldContainSubstring, "emu")
			So(m.Messages, ShouldContainSubstring, "green")

			Infof("pig", "purple")
			So(m.Messages, ShouldContainSubstring, "pig")
			So(m.Messages, ShouldContainSubstring, "purple")

			Warning("jellyfish", "yellow")
			So(m.Messages, ShouldContainSubstring, "jellyfish")
			So(m.Messages, ShouldContainSubstring, "yellow")

			Warningf("lion", "white")
			So(m.Messages, ShouldContainSubstring, "lion")
			So(m.Messages, ShouldContainSubstring, "white")

			Error("monkey", "grey")
			So(m.Messages, ShouldContainSubstring, "monkey")
			So(m.Messages, ShouldContainSubstring, "grey")

			Errorf("parrot%s", "black")
			So(m.Messages, ShouldContainSubstring, "parrot")
			So(m.Messages, ShouldContainSubstring, "black")

			Fatal("spider", "cyan")
			So(m.Messages, ShouldContainSubstring, "spider")
			So(m.Messages, ShouldContainSubstring, "cyan")

			Fatalf("ladybug%s", "pink")
			So(m.Messages, ShouldContainSubstring, "ladybug")
			So(m.Messages, ShouldContainSubstring, "pink")
		})
		Convey("info level", func() {
			m := &MockWriter{}
			Output = m
			logger = log.New(Output, "", log.LstdFlags)
			SetLevel(InfoLevel)

			Debug("sheep", "red")
			So(m.Messages, ShouldNotContainSubstring, "sheep")
			So(m.Messages, ShouldNotContainSubstring, "red")

			Debugf("cow", "blue")
			So(m.Messages, ShouldNotContainSubstring, "cow")
			So(m.Messages, ShouldNotContainSubstring, "blue")

			Info("emu", "green")
			So(m.Messages, ShouldContainSubstring, "emu")
			So(m.Messages, ShouldContainSubstring, "green")

			Infof("pig", "purple")
			So(m.Messages, ShouldContainSubstring, "pig")
			So(m.Messages, ShouldContainSubstring, "purple")

			Warning("jellyfish", "yellow")
			So(m.Messages, ShouldContainSubstring, "jellyfish")
			So(m.Messages, ShouldContainSubstring, "yellow")

			Warningf("lion", "white")
			So(m.Messages, ShouldContainSubstring, "lion")
			So(m.Messages, ShouldContainSubstring, "white")

			Error("monkey", "grey")
			So(m.Messages, ShouldContainSubstring, "monkey")
			So(m.Messages, ShouldContainSubstring, "grey")

			Errorf("parrot%s", "black")
			So(m.Messages, ShouldContainSubstring, "parrot")
			So(m.Messages, ShouldContainSubstring, "black")

			Fatal("spider", "cyan")
			So(m.Messages, ShouldContainSubstring, "spider")
			So(m.Messages, ShouldContainSubstring, "cyan")

			Fatalf("ladybug%s", "pink")
			So(m.Messages, ShouldContainSubstring, "ladybug")
			So(m.Messages, ShouldContainSubstring, "pink")
		})
		Convey("warning level", func() {
			m := &MockWriter{}
			Output = m
			logger = log.New(Output, "", log.LstdFlags)
			SetLevel(WarningLevel)

			Debug("sheep", "red")
			So(m.Messages, ShouldNotContainSubstring, "sheep")
			So(m.Messages, ShouldNotContainSubstring, "red")

			Debugf("cow", "blue")
			So(m.Messages, ShouldNotContainSubstring, "cow")
			So(m.Messages, ShouldNotContainSubstring, "blue")

			Info("emu", "green")
			So(m.Messages, ShouldNotContainSubstring, "emu")
			So(m.Messages, ShouldNotContainSubstring, "green")

			Infof("pig", "purple")
			So(m.Messages, ShouldNotContainSubstring, "pig")
			So(m.Messages, ShouldNotContainSubstring, "purple")

			Warning("jellyfish", "yellow")
			So(m.Messages, ShouldContainSubstring, "jellyfish")
			So(m.Messages, ShouldContainSubstring, "yellow")

			Warningf("lion", "white")
			So(m.Messages, ShouldContainSubstring, "lion")
			So(m.Messages, ShouldContainSubstring, "white")

			Error("monkey", "grey")
			So(m.Messages, ShouldContainSubstring, "monkey")
			So(m.Messages, ShouldContainSubstring, "grey")

			Errorf("parrot%s", "black")
			So(m.Messages, ShouldContainSubstring, "parrot")
			So(m.Messages, ShouldContainSubstring, "black")

			Fatal("spider", "cyan")
			So(m.Messages, ShouldContainSubstring, "spider")
			So(m.Messages, ShouldContainSubstring, "cyan")

			Fatalf("ladybug%s", "pink")
			So(m.Messages, ShouldContainSubstring, "ladybug")
			So(m.Messages, ShouldContainSubstring, "pink")
		})
		Convey("error level", func() {
			m := &MockWriter{}
			Output = m
			logger = log.New(Output, "", log.LstdFlags)
			SetLevel(ErrorLevel)

			Debug("sheep", "red")
			So(m.Messages, ShouldNotContainSubstring, "sheep")
			So(m.Messages, ShouldNotContainSubstring, "red")

			Debugf("cow", "blue")
			So(m.Messages, ShouldNotContainSubstring, "cow")
			So(m.Messages, ShouldNotContainSubstring, "blue")

			Info("emu", "green")
			So(m.Messages, ShouldNotContainSubstring, "emu")
			So(m.Messages, ShouldNotContainSubstring, "green")

			Infof("pig", "purple")
			So(m.Messages, ShouldNotContainSubstring, "pig")
			So(m.Messages, ShouldNotContainSubstring, "purple")

			Warning("jellyfish", "yellow")
			So(m.Messages, ShouldNotContainSubstring, "jellyfish")
			So(m.Messages, ShouldNotContainSubstring, "yellow")

			Warningf("lion", "white")
			So(m.Messages, ShouldNotContainSubstring, "lion")
			So(m.Messages, ShouldNotContainSubstring, "white")

			Error("monkey", "grey")
			So(m.Messages, ShouldContainSubstring, "monkey")
			So(m.Messages, ShouldContainSubstring, "grey")

			Errorf("parrot%s", "black")
			So(m.Messages, ShouldContainSubstring, "parrot")
			So(m.Messages, ShouldContainSubstring, "black")

			Fatal("spider", "cyan")
			So(m.Messages, ShouldContainSubstring, "spider")
			So(m.Messages, ShouldContainSubstring, "cyan")

			Fatalf("ladybug%s", "pink")
			So(m.Messages, ShouldContainSubstring, "ladybug")
			So(m.Messages, ShouldContainSubstring, "pink")
		})
		Convey("fatal level", func() {
			m := &MockWriter{}
			Output = m
			logger = log.New(Output, "", log.LstdFlags)
			SetLevel(FatalLevel)

			Debug("sheep", "red")
			So(m.Messages, ShouldNotContainSubstring, "sheep")
			So(m.Messages, ShouldNotContainSubstring, "red")

			Debugf("cow", "blue")
			So(m.Messages, ShouldNotContainSubstring, "cow")
			So(m.Messages, ShouldNotContainSubstring, "blue")

			Info("emu", "green")
			So(m.Messages, ShouldNotContainSubstring, "emu")
			So(m.Messages, ShouldNotContainSubstring, "green")

			Infof("pig", "purple")
			So(m.Messages, ShouldNotContainSubstring, "pig")
			So(m.Messages, ShouldNotContainSubstring, "purple")

			Warning("jellyfish", "yellow")
			So(m.Messages, ShouldNotContainSubstring, "jellyfish")
			So(m.Messages, ShouldNotContainSubstring, "yellow")

			Warningf("lion", "white")
			So(m.Messages, ShouldNotContainSubstring, "lion")
			So(m.Messages, ShouldNotContainSubstring, "white")

			Error("monkey", "grey")
			So(m.Messages, ShouldNotContainSubstring, "monkey")
			So(m.Messages, ShouldNotContainSubstring, "grey")

			Errorf("parrot%s", "black")
			So(m.Messages, ShouldNotContainSubstring, "parrot")
			So(m.Messages, ShouldNotContainSubstring, "black")

			Fatal("spider", "cyan")
			So(m.Messages, ShouldContainSubstring, "spider")
			So(m.Messages, ShouldContainSubstring, "cyan")

			Fatalf("ladybug%s", "pink")
			So(m.Messages, ShouldContainSubstring, "ladybug")
			So(m.Messages, ShouldContainSubstring, "pink")
		})
	})
}
