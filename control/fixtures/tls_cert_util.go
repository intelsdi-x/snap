// +build legacy small medium large

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package fixtures

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	mrand "math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	TestCrtFileExt    = ".crt"
	TestBadCrtFileExt = "-BAD.crt"
	TestKeyFileExt    = ".key"
)

const (
	keyBitsDefault            = 2048
	defaultKeyValidPeriod     = 6 * time.Hour
	rsaKeyPEMHeader           = "RSA PRIVATE KEY"
	certificatePEMHeader      = "CERTIFICATE"
	defaultSignatureAlgorithm = x509.SHA256WithRSA
	defaultPublicKeyAlgorithm = x509.RSA
)

// CertTestUtil offers a few methods to generate a few self-signed certificates
// suitable only for test.
type CertTestUtil struct {
	Prefix string
}

// WritePEMFile writes block of bytes into a PEM formatted file with given header.
func (u CertTestUtil) WritePEMFile(fn string, pemHeader string, b []byte) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	pem.Encode(w, &pem.Block{
		Type:  pemHeader,
		Bytes: b,
	})
	w.Flush()
	return nil
}

// MakeCACertKeyPair generates asymmetric private key and certificate
// for CA, suitable for signing certificates
func (u CertTestUtil) MakeCACertKeyPair(caName, ouName string, keyValidPeriod time.Duration) (caCertTpl *x509.Certificate, caCertBytes []byte, caPrivKey *rsa.PrivateKey, err error) {
	caPrivKey, err = rsa.GenerateKey(rand.Reader, keyBitsDefault)
	if err != nil {
		return nil, nil, nil, err
	}
	caPubKey := caPrivKey.Public()
	caPubBytes, err := x509.MarshalPKIXPublicKey(caPubKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caPubSha256 := sha256.Sum256(caPubBytes)
	caCertTpl = &x509.Certificate{
		SignatureAlgorithm: defaultSignatureAlgorithm,
		PublicKeyAlgorithm: defaultPublicKeyAlgorithm,
		Version:            3,
		SerialNumber:       big.NewInt(1),
		Subject: pkix.Name{
			CommonName: caName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(keyValidPeriod),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
		IsCA:                  true,
		SubjectKeyId:          caPubSha256[:],
	}
	caCertBytes, err = x509.CreateCertificate(rand.Reader, caCertTpl, caCertTpl, caPubKey, caPrivKey)
	if err != nil {
		return nil, nil, nil, err
	}
	return caCertTpl, caCertBytes, caPrivKey, nil
}

// MakeSubjCertKeyPair generates a private key and a certificate for subject
// suitable for securing TLS communication
func (u CertTestUtil) MakeSubjCertKeyPair(cn, ou string, keyValidPeriod time.Duration, caCertTpl *x509.Certificate, caPrivKey *rsa.PrivateKey) (subjCertBytes []byte, subjPrivKey *rsa.PrivateKey, err error) {
	subjPrivKey, err = rsa.GenerateKey(rand.Reader, keyBitsDefault)
	if err != nil {
		return nil, nil, err
	}
	subjPubBytes, err := x509.MarshalPKIXPublicKey(subjPrivKey.Public())
	if err != nil {
		return nil, nil, err
	}
	subjPubSha256 := sha256.Sum256(subjPubBytes)
	subjCertTpl := x509.Certificate{
		SignatureAlgorithm: defaultSignatureAlgorithm,
		PublicKeyAlgorithm: defaultPublicKeyAlgorithm,
		Version:            3,
		SerialNumber:       big.NewInt(1),
		Subject: pkix.Name{
			OrganizationalUnit: []string{ou},
			CommonName:         cn,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(keyValidPeriod),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyAgreement,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		SubjectKeyId: subjPubSha256[:],
	}
	subjCertTpl.DNSNames = []string{"localhost"}
	subjCertTpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	subjCertBytes, err = x509.CreateCertificate(rand.Reader, &subjCertTpl, caCertTpl, subjPrivKey.Public(), caPrivKey)
	return subjCertBytes, subjPrivKey, err
}

// StoreTLSCerts builds a set of certificates and private keys for testing TLS.
// Generated files include: CA certificate, server certificate and private key,
// client certificate and private key, and alternate (BAD) CA certificate.
// Certificate and key files are named after given common names (e.g.: srvCN).
func (u CertTestUtil) StoreTLSCerts(caCN, srvCN, cliCN string) (resFiles []string, err error) {
	ou := fmt.Sprintf("%06x", mrand.Intn(1<<24))
	caCertTpl, caCert, caPrivKey, err := u.MakeCACertKeyPair(caCN, ou, defaultKeyValidPeriod)
	if err != nil {
		return nil, err
	}
	caCertFn := filepath.Join(u.Prefix, caCN+TestCrtFileExt)
	if err := u.WritePEMFile(caCertFn, certificatePEMHeader, caCert); err != nil {
		return nil, err
	}
	resFiles = append(resFiles, caCertFn)
	_, caBadCert, _, err := u.MakeCACertKeyPair(caCN, ou, defaultKeyValidPeriod)
	if err != nil {
		return resFiles, err
	}
	badCaCertFn := caCN + TestBadCrtFileExt
	if err := u.WritePEMFile(badCaCertFn, certificatePEMHeader, caBadCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, badCaCertFn)
	srvCert, srvPrivKey, err := u.MakeSubjCertKeyPair(srvCN, ou, defaultKeyValidPeriod, caCertTpl, caPrivKey)
	if err != nil {
		return resFiles, err
	}
	srvCertFn := filepath.Join(u.Prefix, srvCN+TestCrtFileExt)
	srvKeyFn := filepath.Join(u.Prefix, srvCN+TestKeyFileExt)
	if err := u.WritePEMFile(srvCertFn, certificatePEMHeader, srvCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, srvCertFn)
	if err := u.WritePEMFile(srvKeyFn, rsaKeyPEMHeader, x509.MarshalPKCS1PrivateKey(srvPrivKey)); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, srvKeyFn)
	cliCert, cliPrivKey, err := u.MakeSubjCertKeyPair(cliCN, ou, defaultKeyValidPeriod, caCertTpl, caPrivKey)
	if err != nil {
		return resFiles, err
	}
	cliCertFn := filepath.Join(u.Prefix, cliCN+TestCrtFileExt)
	cliKeyFn := filepath.Join(u.Prefix, cliCN+TestKeyFileExt)
	if err := u.WritePEMFile(cliCertFn, certificatePEMHeader, cliCert); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, cliCertFn)
	if err := u.WritePEMFile(cliKeyFn, rsaKeyPEMHeader, x509.MarshalPKCS1PrivateKey(cliPrivKey)); err != nil {
		return resFiles, err
	}
	resFiles = append(resFiles, cliKeyFn)
	return resFiles, nil
}
