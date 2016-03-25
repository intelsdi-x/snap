#!/bin/bash

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#Copyright 2015 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

# add some color to the output
red=`tput setaf 1`
green=`tput setaf 2`
reset=`tput sgr0`

die () {
    echo >&2 "${red} $@ ${reset}"
    exit 1
}

type cfssl >/dev/null 2>&1 && type cfssljson >/dev/null 2>&1 || die "Error: cfssl and cfssljson are required (see https://github.com/cloudflare/cfssl)"

# Create the CA
cfssl gencert -initca config/ca-csr.json | cfssljson -bare sample-ca

# Generate a signed cert
cfssl gencert \
  -ca=sample-ca.pem \
  -ca-key=sample-ca-key.pem \
  -config=config/config.json \
  -profile=cert \
  config/cert-csr.json | cfssljson -bare sample-signed-cert

# Generate a unsigned cert
cfssl gencert \
  -config=config/config.json \
  -profile=cert \
  config/cert-csr.json | cfssljson -bare sample-unsigned-cert


