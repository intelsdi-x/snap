package psigning

import (
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSignature(t *testing.T) {
	keyringFile := "pubring.gpg"
	signedFile := "pulse-collector-dummy1"
	signatureFile := signedFile + ".asc"
	pulsePath := os.Getenv("PULSE_PATH")
	unsignedFile := path.Join(pulsePath, "plugin", "pulse-collector-dummy2")

	s := SigningManager{}

	Convey("Valid files and good signature", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, signatureFile)
		So(err, ShouldBeNil)
	})

	Convey("Valid files and bad signature", t, func() {
		err := s.ValidateSignature(keyringFile, unsignedFile, signatureFile)
		So(err.Error(), ShouldResemble, "Error checking signature")
	})

	Convey("Invalid keyring", t, func() {
		err := s.ValidateSignature("", signedFile, signatureFile)
		So(err.Error(), ShouldResemble, "Keyring file (.gpg) not found")
	})

	Convey("Invalid signed file", t, func() {
		err := s.ValidateSignature(keyringFile, "", signatureFile)
		So(err.Error(), ShouldResemble, "Signed file not found")
	})

	Convey("Invalid signature file", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, "")
		So(err.Error(), ShouldResemble, "Signature file (.asc) not found")
	})
}
