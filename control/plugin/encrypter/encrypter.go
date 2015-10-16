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

package encrypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"hash"
	"io"
	"io/ioutil"
)

var ErrKeyNotValid = errors.New("given key length is invalid. did you set it?")

const (
	nonceSize = 12
	keySize   = 32
)

func GenerateKey() ([]byte, error) {
	key := make([]byte, keySize)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

type Encrypter struct {
	Key []byte

	rsaPublic  *rsa.PublicKey
	rsaPrivate *rsa.PrivateKey
	md5        hash.Hash
}

func New(pub *rsa.PublicKey, priv *rsa.PrivateKey) *Encrypter {
	return &Encrypter{
		rsaPublic:  pub,
		rsaPrivate: priv,
		md5:        md5.New(),
	}
}

func (e *Encrypter) Encrypt(in io.Reader) ([]byte, error) {
	if len(e.Key) < keySize {
		return nil, ErrKeyNotValid
	}

	bytes, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher(e.Key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, bytes, nil), nil
}

func (e *Encrypter) Decrypt(in io.Reader) ([]byte, error) {
	if len(e.Key) < keySize {
		return nil, ErrKeyNotValid
	}

	bytes, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher(e.Key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce := bytes[:nonceSize]
	return gcm.Open(nil, nonce, bytes[nonceSize:], nil)
}

func (e *Encrypter) EncryptKey() ([]byte, error) {
	if len(e.Key) != keySize {
		return nil, ErrKeyNotValid
	}
	return rsa.EncryptOAEP(e.md5, rand.Reader, e.rsaPublic, e.Key, []byte(""))
}

func (e *Encrypter) DecryptKey(in []byte) ([]byte, error) {
	return rsa.DecryptOAEP(e.md5, rand.Reader, e.rsaPrivate, in, []byte(""))
}

func generateNonce() ([]byte, error) {
	nonce := make([]byte, nonceSize)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}
