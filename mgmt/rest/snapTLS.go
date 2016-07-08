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

package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

type snapTLS struct {
	cert, key string
}

func newtls(certPath, keyPath string) (*snapTLS, error) {
	t := &snapTLS{}
	if certPath != "" && keyPath != "" {
		cert, err := os.Open(certPath)
		if err != nil {
			return nil, err
		}
		_, err = os.Open(keyPath)
		if err != nil {
			return nil, err
		}

		rest, err := ioutil.ReadAll(cert)
		if err != nil {
			return nil, err
		}

		// test that given keychain is valid
		var block *pem.Block
		for {
			block, rest = pem.Decode(rest)
			if block == nil {
				return nil, ErrBadCert
			}
			_, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			if len(rest) == 0 {
				break
			}
		}
		t.cert = certPath
		t.key = keyPath
	} else {
		err := generateCert(t)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func generateCert(t *snapTLS) error {
	// good for 1 year
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 365)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	temp := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Local snap Agent"},
		},
		DNSNames:              []string{"localhost"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	cbytes, err := x509.CreateCertificate(rand.Reader, temp, temp, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certPath := os.TempDir() + "/cert.pem"
	keyPath := os.TempDir() + "/key.pem"

	cout, err := os.Create(certPath)
	if err != nil {
		return err
	}
	pem.Encode(cout, &pem.Block{Type: "CERTIFICATE", Bytes: cbytes})
	cout.Close()

	kout, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(kout, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	kout.Close()

	t.cert = certPath
	t.key = keyPath

	return nil
}
