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
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/openpgp"
)

type SigningManager struct{}

var (
	ErrPluginNotFound        = errors.New("Plugin not found")
	ErrKeyringFileNotFound   = errors.New("Keyring file (.gpg) not found")
	ErrUnableToReadKeyring   = errors.New("Unable to read keyring")
	ErrSignedFileNotFound    = errors.New("Signed file not found")
	ErrSignatureFileNotFound = errors.New("Signature file (.asc) not found. Did you use the -a flag?")
	ErrCheckSignature        = errors.New("Error checking signature")
)

//ValidateSignature is exported for plugin authoring
func (s *SigningManager) ValidateSignature(keyringFile, signedFile, signatureFile string) error {
	keyringf, err := os.Open(keyringFile)
	if err != nil {
		return fmt.Errorf("%v: %v\n%v", ErrKeyringFileNotFound, keyringFile, err)
	}
	defer keyringf.Close()

	keyring, err := openpgp.ReadKeyRing(keyringf)
	if err != nil {
		return fmt.Errorf("%v: %v\n%v", ErrUnableToReadKeyring, keyringFile, err)
	}

	signed, err := os.Open(signedFile)
	if err != nil {
		return fmt.Errorf("%v: %v\n%v", ErrSignedFileNotFound, signedFile, err)
	}
	defer signed.Close()

	signature, err := os.Open(signatureFile)
	if err != nil {
		return fmt.Errorf("%v: %v\n%v", ErrSignatureFileNotFound, signatureFile, err)
	}
	defer signature.Close()

	checked, err := openpgp.CheckArmoredDetachedSignature(keyring, signed, signature)
	if err != nil {
		return fmt.Errorf("%v\n%v", ErrCheckSignature, err)
	}

	var signedby string
	for k := range checked.Identities {
		signedby = signedby + k
	}
	fmt.Printf("Signature made %v using RSA key ID %v\nGood signature from %v\n", time.Now().Format(time.RFC1123), checked.PrimaryKey.KeyIdShortString(), signedby)
	return nil
}
