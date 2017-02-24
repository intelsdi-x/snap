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
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GetClientConnection returns a grcp.ClientConn that is unsecured
// TODO: Add TLS security to connection
func GetClientConnection(addr string, port int) (*grpc.ClientConn, error) {
	return GetClientConnectionWithCreds(addr, port, nil)
}

func GetClientConnectionWithCreds(addr string, port int, creds *credentials.TransportCredentials) (*grpc.ClientConn, error) {
	grpcDialOpts := []grpc.DialOption{
		grpc.WithTimeout(2 * time.Second),
	}
	if creds != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(*creds))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", addr, port), grpcDialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
