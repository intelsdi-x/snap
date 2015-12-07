/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package psigning

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/openpgp"
)

// SigningManager anonymous struct type
type SigningManager struct{}

var (
	// ErrPluginNotFound - Error message for plugin not found
	ErrPluginNotFound = errors.New("Plugin not found")
	// ErrKeyringFileNotFound - Error message for keyring file (.gpg) not found
	ErrKeyringFileNotFound = errors.New("Keyring file (.gpg) not found")
	// ErrUnableToReadKeyring - Error message for unable to read keyring
	ErrUnableToReadKeyring = errors.New("Unable to read keyring")
	// ErrSignedFileNotFound - Error message for signed file not found
	ErrSignedFileNotFound = errors.New("Signed file not found")
	// ErrCheckSignature - Error message for error checking signature
	ErrCheckSignature = errors.New("Error checking signature")
)

//ValidateSignature is exported for plugin authoring
func (s *SigningManager) ValidateSignature(keyringFiles []string, signedFile string, signature []byte) error {
	var signedby string
	var e error
	var checked *openpgp.Entity

	signed, err := os.Open(signedFile)
	if err != nil {
		return fmt.Errorf("%v: %v\n%v", ErrSignedFileNotFound, signedFile, err)
	}
	defer signed.Close()

	//Go through all the keyrings til either signature is valid or end of keyrings
	for _, keyringFile := range keyringFiles {
		keyringf, err := os.Open(keyringFile)
		if err != nil {
			return fmt.Errorf("%v: %v\n%v", ErrKeyringFileNotFound, keyringFile, err)
		}
		defer keyringf.Close()

		//Read both armored and unarmored keyrings
		keyring, err := openpgp.ReadArmoredKeyRing(keyringf)
		if err != nil {
			keyringf.Seek(0, 0)
			keyring, err = openpgp.ReadKeyRing(keyringf)
			if err != nil {
				return fmt.Errorf("%v: %v\n%v", ErrUnableToReadKeyring, keyringFile, err)
			}
		}

		//Check the armored detached signature
		checked, e = openpgp.CheckArmoredDetachedSignature(keyring, signed, bytes.NewReader(signature))
		if e == nil {
			for k := range checked.Identities {
				signedby = signedby + k
			}
			fmt.Printf("Signature made %v using RSA key ID %v\nGood signature from %v\n", time.Now().Format(time.RFC1123), checked.PrimaryKey.KeyIdShortString(), signedby)
			return nil
		}
		signed.Seek(0, 0)
	}
	return fmt.Errorf("%v\n%v", ErrCheckSignature, e)
}
