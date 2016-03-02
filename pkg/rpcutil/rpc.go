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

package rpcutil

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	// ErrCertTrust is returned when a a non CA cert is provided to GrpcTlsOption
	ErrCertTrust = errors.New("A trusted root certificate (CA) is required")
)

// ServerTlsOption returns a grpc.ServerOption enabling tls.  A server cert is generated and signed
// with the provided CA.  The server will require and verify a client cert.
func ServerTlsOption(caCert, caKey, listenAddress string) (grpc.ServerOption, error) {
	crt, err := genServerCert(caCert, caKey, listenAddress)
	if err != nil {
		return nil, err
	}
	cac, err := ioutil.ReadFile(caCert)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cac)
	ta := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	return grpc.Creds(ta), nil

}

// GetClientConnection returns a grcp.ClientConn that is secured (tls) if a caCert and caKey
// is provided.  If the caCert and caKey are empty the grpc communication is not secured.
func GetClientConnection(addr string, port int, caCert, caKey string) (*grpc.ClientConn, error) {
	grpcDialOpts := []grpc.DialOption{
		grpc.WithTimeout(2 * time.Second),
	}
	if caCert != "" && caKey != "" {
		clientCert, err := genClientCert(caCert, caKey)
		if err != nil {
			return nil, err
		}
		caCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		ta := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      caCertPool,
		})
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(ta))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", addr, port), grpcDialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func genServerCert(caCert, caKey, listAddress string) (tls.Certificate, error) {
	return genCert(caCert, caKey, listAddress)
}

func genClientCert(caCert, caKey string) (tls.Certificate, error) {
	return genCert(caCert, caKey, "")
}

func genCert(caCert, caKey, listenAddress string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(caCert, caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	ca, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}
	if !ca.IsCA {
		return tls.Certificate{}, ErrCertTrust
	}

	notBefore := time.Now().Add(-time.Hour * 24 * 7)     // 1 week ago
	notAfter := notBefore.Add(time.Hour * 24 * 365 * 10) // 10 years
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	temp := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Local snap Agent"},
		},
		DNSNames:                    nil,
		PermittedDNSDomainsCritical: false,
		PermittedDNSDomains:         nil,
		NotBefore:                   notBefore,
		NotAfter:                    notAfter,
		KeyUsage:                    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:                 []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid:       true,
	}
	if listenAddress != "" {
		temp.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP(listenAddress)}
	}

	k, err := ioutil.ReadFile(caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	block, _ := pem.Decode([]byte(k))
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return tls.Certificate{}, err
	}

	cbytes, err := x509.CreateCertificate(rand.Reader, temp, ca, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var cout bytes.Buffer
	c := bufio.NewWriter(&cout)
	pem.Encode(c, &pem.Block{Type: "CERTIFICATE", Bytes: cbytes})
	if err != nil {
		return tls.Certificate{}, err
	}
	c.Flush()

	crt, err := tls.X509KeyPair(cout.Bytes(), k)
	if err != nil {
		return tls.Certificate{}, err
	}

	return crt, nil
}
