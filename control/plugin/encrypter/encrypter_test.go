package encrypter

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEncrypter(t *testing.T) {
	Convey("Encrypter", t, func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		So(err, ShouldBeNil)
		e := New(&key.PublicKey, key)
		symkey, err := GenerateKey()
		So(err, ShouldBeNil)
		e.Key = symkey
		Convey("The constructor works", func() {
			So(e, ShouldHaveSameTypeAs, &Encrypter{})
		})
		Convey("it can encrypt stuff", func() {
			_, err := e.Encrypt(bytes.NewReader([]byte("hello, encrypter")))
			So(err, ShouldBeNil)
		})
		Convey("it can decrypt stuff", func() {
			out, err := e.Encrypt(bytes.NewReader([]byte("hello, encrypter")))
			So(err, ShouldBeNil)
			dec, err := e.Decrypt(bytes.NewReader(out))
			So(err, ShouldBeNil)
			So(string(dec), ShouldEqual, "hello, encrypter")
		})
	})
}
