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

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// grpcDialDefaultTimeout is the default timeout for initial gRPC dial
const grpcDialDefaultTimeout = 2 * time.Second

// GetClientConnection returns a grcp.ClientConn that is unsecured
func GetClientConnection(ctx context.Context, addr string, port int) (*grpc.ClientConn, error) {
	return GetClientConnectionWithCreds(ctx, addr, port, nil)
}

// GetClientConnectionWithCreds returns a grcp.ClientConn with optional TLS
// security (if creds != nil)
func GetClientConnectionWithCreds(ctx context.Context, addr string, port int, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	grpcDialOpts := []grpc.DialOption{
		grpc.WithTimeout(grpcDialDefaultTimeout),
	}
	if creds != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(creds))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%v:%v", addr, port), grpcDialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
