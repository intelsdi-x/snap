<!--
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
-->

# Secure Plugin Communication

<!-- TOC -->

- [Secure Plugin Communication](#secure-plugin-communication)
    - [Overview](#overview)
    - [Prerequisites](#prerequisites)
    - [Plugin modes of operation allowing TLS encrypted communication](#plugin-modes-of-operation-allowing-tls-encrypted-communication)
        - [Stand-alone plugins running on web addressable hosts](#stand-alone-plugins-running-on-web-addressable-hosts)
        - [Plugins running locally](#plugins-running-locally)
    - [Detailed preparation](#detailed-preparation)
        - [Enabling secure communication](#enabling-secure-communication)
        - [Using system-installed CA certificates](#using-system-installed-ca-certificates)
    - [More information](#more-information)
        - [Exclusive security](#exclusive-security)
        - [Relation to other functionalities](#relation-to-other-functionalities)
        - [TLS setup requirements](#tls-setup-requirements)
        - [Obtaining self-signed TLS certificates for tests](#obtaining-self-signed-tls-certificates-for-tests)
    - [More information](#more-information-1)

<!-- /TOC -->

## Overview

Snap framework communicates with plugins over gRPC protocol, which in general transfers data in plaintext.
Snap allows securing communication with plugins by opening TLS channels and using certificates to authenticate plugins and framework.

Instructions given in this document allows launching workflows with TLS encrypted communication.

## Prerequisites

This walkthrough assumes you have downloaded a Snap release as described in [Getting Started](../README.md#getting-started).

## Plugin modes of operation allowing TLS encrypted communication

At present there are two modes of plugin operation that support TLS secured communication:
* stand-alone plugins running on web addressable hosts,
* locally launched plugins.

### Stand-alone plugins running on web addressable hosts

TLS may be employed to secure communication with stand-alone plugins. In such setup the Snap will be communicating securely with remote plugin instances running in dedicated domains.

Assuming that TLS certificate and key files are available (see [SETUP_TLS_CERTIFICATES](SETUP_TLS_CERTIFICATES.md)), the following steps will result in plugin communication encrypted with TLS:

On a remote host mapped to own domain, e.g. `collectors.example.net`:
```bash
$ snap-plugin-collector-psutil --stand-alone --stand-alone-port 8281 --addr collectors.example.net --tls --cert-path /tmp/collectors.example.net-srv.crt --key-path /tmp/collectors.example.net-srv.key --root-cert-paths /tmp/example.net-ca.crt
$ snap-plugin-publisher-influxdb --stand-alone --stand-alone-port 8381 --addr collectors.example.net --tls --cert-path /tmp/collectors.example.net-srv.crt --key-path /tmp/collectors.example.net-srv.key --root-cert-paths /tmp/example.net-ca.crt
```

On a machine dedicated for running framework daemon `snapteld`:
```bash
$ snapteld --log-level 1 --plugin-trust 0 --tls-cert=/tmp/daemon.example.net-cli.crt --tls-key=/tmp/daemon.example.net-cli.key --ca-cert-paths=/tmp/example.net-ca.crt
$ snaptel plugin load http://collectors.example.net:8281
$ snaptel plugin load http://collectors.example.net:8381
$ snaptel task create -t sample-task.json
```

### Plugins running locally

Assuming all the test files are available (basing on [test instructions](#obtaining-self-signed-tls-certificates-for-tests)), the following steps will result in plugin communication encrypted with TLS:

```bash
$ snapteld --log-level 1 --plugin-trust 0 --tls-cert /tmp/snaptest-cli.crt --tls-key /tmp/snaptest-cli.key --ca-cert-paths /tmp/snaptest-ca.crt
$ snaptel plugin load --plugin-cert /tmp/snaptest-srv.crt --plugin-key /tmp/snaptest-srv.key --plugin-ca-certs /tmp/snaptest-ca.crt plugins/snap-plugin-collector-rand
$ snaptel task create -t sample-task.json
```

## Detailed preparation

Starting secure communication requires following steps:
1. Obtain X.509 certificate and private key for framework.
    * Please note that this certificate should allow usage for TLS web client authentication (as specified in RFC 3280)
1. Obtain X.509 certificate and private key for each plugin or group of plugins.
    * Please note that this certificate should allow usage for TLS web server authentication (as specified in RFC 3280)
1. Obtain and locate the CA certificates that are necessary to authenticate framework and plugin certificates.

Process of acquiring a TLS certificate is a complex one. Every organization has its specific rules on security, thus the details are not given here.

We do provide a short guide on obtaining self-signed certificates that may be used for tests outside production environment; see [Obtaining self-signed TLS certificates for tests](#obtaining-self-signed-tls-certificates-for-tests).

### Enabling secure communication

Secure communication is enabled by passing the required paths to programs: `snapteld`, and plugin (via `snaptel`). The minimum paths necessary are:
* for `snapteld`: `--tls-cert`, `--tls-key`
* for locally running plugins (`snaptel`): `--plugin-cert`, `--plugin-key`
* for plugins started in stand-alone mode: `--tls`, `--cert-path`, `--key-path`.

The required paths are sufficient and necessary to enable TLS. Daemon (`snapteld`) and plugins (via `snaptel`) will refuse to start if certificate or key file argument is missing.

### Using system-installed CA certificates

Framework and plugins need CA certificates to validate each other's certificate. The CA certificates may be obtained in two ways:
* by passing a list of CA certificate paths directly as a parameter, e.g.: `--ca-cert-paths=/tmp/small-setup-ca.crt:/tmp/medium-setup-ca.crt:/tmp/ca-certs/` (snapteld), `--plugin-ca-certs` (snaptel), `--root-cert-paths` (stand-alone plugin)
    * plugin as well as framework will examine each path in the list and either load a file directly or list directory contents and load the enumerated files (e.g.: the files in `/tmp/ca-certs/` folder) 
* by relying on default CA certificate discovery mechanism
    * plugin and framework will by default load certificates from system (if no paths were given as parameter). Each OS has its own specific locations, e.g.: `/etc/ssl/certs` on Ubuntu. This mechanism is provided by Go language, and is only available on selected OSes.     
    System CA certificates may also be loaded explicitly by listing system locations explicitly, e.g.: `--ca-cert-paths /etc/ssl/certs:/tmp/snaptest-ca.crt`

## More information

### Exclusive security

It's important to note that once secure plugin communication is enabled in framework, only secure connections may be established. 

In other words: attempting to load insecure plugin in framework will result in an error. 

### Relation to other functionalities

Several modes of operation do not fully support secure communication:
* distributed workflow is not covered by secure communication,
* tribe doesn't support secure communication; `snapteld` will refuse to start in tribe mode if configured with secure communication,
* plugin and task autodiscovery doesn't support secure communication; `snapteld` will refuse to start with autodiscovery path and secure communication enabled.

### TLS setup requirements

Snap plugin security is subject to following constraints:
* certificates must be valid for use with following cipher suites:
    * TLS_RSA_WITH_AES_128_GCM_SHA256,
    * TLS_RSA_WITH_AES_256_GCM_SHA384.
* certificates should allow usage for TLS web client or server authentication, as specified in RFC 3280 (server usage is for plugins).

### Obtaining self-signed TLS certificates for tests

The following instructions will result in self-signed TLS certificate files. These files may be used for manual tests. For officially signed certificates refer to document [SETUP_TLS_CERTIFICATES](SETUP_TLS_CERTIFICATES.md).
1. Install tool [certstrap](https://github.com/square/certstrap) for generating test certificates. Further steps will assume that `certstrap` is available under `$PATH` location.
1. Generate root CA certificate:
    ```
    certstrap init --cn "snaptest-ca" --o "snaptest" --ou "ca" --key-bits 2048 --years 1 --passphrase ''
    ```
1. **optional** Install root CA certificate in the system:
    ```
    sudo cp out/snaptest-ca.crt /usr/local/share/ca-certificates/; sudo update-ca-certificates --verbose --fresh
    ``` 
1. Generate server certificate and key to use with plugins:
    ```
    certstrap request-cert --cn "snaptest-srv" --ip "127.0.0.1" --domain "localhost" --passphrase '' --key-bits 2048 --o "snaptest" --ou "server"
    certstrap sign "snaptest-srv" --CA "snaptest-ca" --passphrase '' --years 1
    ```
1. Generate client certificate and key to use with `snapteld`:
    ```
    certstrap request-cert --cn "snaptest-cli" --ip "127.0.0.1" --domain "localhost" --passphrase '' --key-bits 2048 --o "snaptest" --ou "client"
    certstrap sign "snaptest-cli" --CA "snaptest-ca" --passphrase '' --years 1
    ```
1. Copy server and client certificates into common location, e.g.: `/tmp`:
    ```
    for unit in srv cli; do for fname in crt key; do cp out/snaptest-$unit.$fname /tmp; done; done
    ```
The following files are relevant for running the tests:
* `/tmp/snaptest-cli.crt`, `/tmp/snaptest-cli.key` - these are the certificate and private key files for `snapteld`,
* `/tmp/snaptest-srv.crt`, `/tmp/snaptest-srv.key` - these are the certificate and private key files for plugin,
* `/tmp/snaptest-ca.crt` - this is the CA certificate that should be used to authenticate framework and plugins.

## More information

* [SETUP_TLS_CERTIFICATES](SETUP_TLS_CERTIFICATES.md)
* [SNAPTELD_CONFIGURATION.md](SNAPTELD_CONFIGURATION.md)
* [SNAPTELD](SNAPTELD.md)
* [SNAPTEL](SNAPTEL.md)
* [TRIBE.md](TRIBE.md)
* [DISTRIBUTED_WORKFLOW_ARCHITECTURE](DISTRIBUTED_WORKFLOW_ARCHITECTURE.md)
