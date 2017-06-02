# Setting up TLS certificates for secure plugin communication in Snap

<!-- TOC -->

- [Setting up TLS certificates for secure plugin communication in Snap](#setting-up-tls-certificates-for-secure-plugin-communication-in-snap)
    - [Overview](#overview)
        - [Remarks](#remarks)
    - [Obtaining self-signed certificates](#obtaining-self-signed-certificates)
        - [Remarks](#remarks-1)
        - [Procedure](#procedure)
    - [Obtaining legitimate certificate from a public CA Let's Encrypt](#obtaining-legitimate-certificate-from-a-public-ca-lets-encrypt)
        - [Prerequisites](#prerequisites)
        - [Remarks](#remarks-2)
        - [Procedure](#procedure-1)
    - [Installing certificates in system (Ubuntu, CentOS)](#installing-certificates-in-system-ubuntu-centos)
        - [Installing root certificate in Ubuntu](#installing-root-certificate-in-ubuntu)
        - [Installing root certificate in Centos](#installing-root-certificate-in-centos)
    - [Testing TLS-secured plugin communication in local setup](#testing-tls-secured-plugin-communication-in-local-setup)
        - [Starting a telemetry workflow:](#starting-a-telemetry-workflow)

<!-- /TOC -->


## Overview

This document provides instructions on how to obtain certficates to evaluate TLS-secured plugin communication in Snap.

Following routines are described:
- obtaining self-signed certificates for evaluation in local test environment,
- obtaining legitimate certificate from a certain CA authority,
- installing root certificate in OS for validating endpoints' certificates,
- starting telemetry workflow with secured plugin communication.

### Remarks
- This document can't be considered a definite guide to obtaining certificates. It aims only as guiding user in acquiring a certificate.
- This guide doesn't indicate how to handle certificate revocation as that service is not yet supported in Snap.
- Every organization has to review its requirements prior to obtaining certificates or setting up any secure communication scheme.

## Obtaining self-signed certificates

This section describes how to obtain certificates signed by a local CA.

In the course of this guide user will establish a local CA, becoming able to establish local chains of trust. Thus the user will be able to generate certificates for Snap plugins and Snap daemon (`snapteld`) that all share common trust root.

### Remarks

- CA private key needs to be protected, possibly on a dedicated machine or media.
- Obtained certificates won't be valid outside local setup.

### Procedure

1. Install tool [certstrap](https://github.com/square/certstrap) for generating test certificates. Further steps will assume that `certstrap` is available under `$PATH` location.
1. Generate root CA certificate:
    ```
    certstrap init --cn "snaptest-ca" --o "snaptest" --ou "ca" --key-bits 2048 --years 1 --passphrase '<ca-secure-passphrase>'
    ```
    Relevant files:
    * `./out/snaptest-ca.key` - CA's private key, encrypted with `<ca-secure-passphrase>`; needs to be protected from unauthorized access,
    * `./out/snaptest-ca.crt` - CA's public certificate, (a certificate root).
1. Generate server certificate and key to use by plugins:
    ```
    certstrap request-cert --cn "snaptest-srv" --ip "127.0.0.1" --domain "localhost" --passphrase '' --key-bits 2048 --o "snaptest" --ou "server"
    certstrap sign "snaptest-srv" --CA "snaptest-ca" --passphrase '<ca-secure-passphrase>' --years 1
    ```
    This will produce a private key certificate file necessary to establish secure TLS channel on server. Server's private key file should not be passphrase-protected (or each plugin would have to ask for passphrase on start).
    Relevant files:
    * `./out/snaptest-srv.key` - server's private key,
    * `./out/snaptest-srv.crt` - server's public certificate.
1. Generate client certificate and key to use by framework:
    ```
    certstrap request-cert --cn "snaptest-cli" --ip "127.0.0.1" --domain "localhost" --passphrase '' --key-bits 2048 --o "snaptest" --ou "client"
    certstrap sign "snaptest-cli" --CA "snaptest-ca" --passphrase '<ca-secure-passphrase>' --years 1
    ```
    This will produce a private key certificate file necessary to securely connect to plugins. Client's private key file should not be passphrase-protected.
    Relevant files:
    * `./out/snaptest-cli.key` - client's private key,
    * `./out/snaptest-cli.crt` - client's public certificate.
    
Files that need to be given at plugin startup:
* `snaptest-ca.crt` - CA's certificate,
* `snaptest-srv.key` - server's private key file,
* `snaptest-srv.crt` - server's certificate file.

Files that need to be given at `snapteld` startup:
* `snaptest-ca.crt` - CA's certificate,
* `snaptest-cli.key` - framework's private key file,
* `snaptest-cli.crt` - framework's certificate file.

## Obtaining legitimate certificate from a public CA Let's Encrypt

This section describes how to obtain a certificate from CA authority. Example will be based on [Let's encrypt](https://letsencrypt.org/) public CA service.

In general, obtaining a TLS certificate involves:
* on customer side: issuing request to CA authority, specifying domain name and contact information,
* on CA side: validating the customer's title to the requested domain.

This guide will show how to obtain certificate for sample domains.

### Prerequisites

There's no way around registering a physical domain. In this walkthrough we'll assume that domains `frodo.buthey.net` and `bilbo.buthey.net` were registered and they're mapped to actual machines (VM or physical).

### Remarks

At the moment there's no way to use domain-bound certificates directly in Snap. That's a subject of ongoing research and design. As soon as this situation changes, this document will be updated.

### Procedure

1. Connect via SSH to a machine mapped to the domain `frodo.buthey.net`.
1. Make sure that machine has WWW ports open for listening: `80` and `443`. 
1. Install [certbot](https://certbot.eff.org/), as directed by Let's encrypt instructions.   
    On Ubuntu:
    ```
    sudo add-apt-repository ppa:certbot/certbot
    sudo apt-get update
    sudo apt-get install certbot
    ``` 
    
    On CentOS:
    ```
    wget https://dl.eff.org/certbot-auto
    chmod a+x certbot-auto
    ``` (CentOS)
1. Request a certificate from Let's Encrypt using Certbot:
    ```
    mkdir -p tmp/certbot/conf.d tmp/certbot/work.d tmp/certbot/log.d
    sudo certbot certonly -d frodo.buthey.net --standalone --config-dir tmp/certbot/conf.d/ --work-dir tmp/certbot/work.d/ --logs-dir tmp/certbot/log.d/ 
    d=frodo.buthey.net; p=$(sudo find . -wholename "*live/$d"); for f in $(sudo ls -1 $p |grep .pem); do sudo cp $p/$f $(basename $p)-$f; done; sudo chown $(whoami) $d-*.pem
    
    ```
    Resulting files:
    * `frodo.buthey.net-privkey.pem` - private key file for domain `frodo.`,
    * `frodo.buthey.net-cert.pem` - certificate file for domain `frodo.`,
    * `frodo.buthey.net-chain.pem` - a certificate chain representing Let's encrypt CA (a certificate root).
    
## Installing certificates in system (Ubuntu, CentOS)

This guide shows how to install root CA certificates in system. Having root CA certificates installed in system reduces complexity of TLS setup.

### Installing root certificate in Ubuntu

1. Obtain a certificate root file (e.g. `snaptest-CA.crt` from the guides in this document).
1. Install the certificate in the system:
    ```
    sudo cp snaptest-ca.crt /usr/local/share/ca-certificates/
    sudo update-ca-certificates --verbose --fresh
    ```
    Sample output:
    ```
    Clearing symlinks in /etc/ssl/certs...
    done.
    Updating certificates in /etc/ssl/certs...
    Doing .
    WARNING: Skipping duplicate certificate ACCVRAIZ1.pem
    WARNING: Skipping duplicate certificate ACCVRAIZ1.pem
    173 added, 0 removed; done.
    ```

### Installing root certificate in Centos

1. Obtain a certificate root file (e.g. `snaptest-ca.crt` from the guides in this document).
1. Install the certificate tooling and certificate in the system:
    ```
    yum install ca-certificates
    update-ca-trust enable
    cp /opt/snap/certs/snaptest-ca.crt /etc/pki/ca-trust/source/anchors/
    update-ca-trust extract
    ```

## Testing TLS-secured plugin communication in local setup

We'll assume the following files are all available:
* `snaptest-ca.crt` - certificate root - installed in system,
* `snaptest-srv.key` - server's private key file,
* `snaptest-srv.crt` - server's certificate file,
* `snaptest-cli.key` - framework's private key file,
* `snaptest-cli.crt` - framework's certificate file.

### Starting a telemetry workflow:

1. Start new shell session and execute `snapteld` with certificate and key generated for client:
    ```
    snapteld -t 0 -l 1 --tls-cert=/tmp/snaptest-cli.crt --tls-key=/tmp/snaptest-cli.key
    ```

1. Load a plugin

    Start new shell session and load plugin built with [snap-plugin-lib-go](https://github.com/intelsdi-x/snap-plugin-lib-go), e.g. `snap-plugin-collector-rand` from `snap-plugin-lib-go` examples:
    ```
    snaptel plugin load --plugin-cert=/tmp/snaptest-srv.crt --plugin-key=/tmp/snaptest-srv.key snap-plugin-collector-rand
    ```
    
1. Start a sample task and watch the output:
    ```
    snaptel task create -t sample-task.json
    
    snaptel task watch 167384ad-7643-43ef-930a-18455f5307a0
    ```
    Sample output:
    ```
    snaptel task create -t task0.json                                                                                                          
    Task created
    ID: 167384ad-7643-43ef-930a-18455f5307a0
    Name: Task-167384ad-7643-43ef-930a-18455f5307a0
    State: Running
    snaptel task watch 167384ad-7643-43ef-930a-18455f5307a0
    Watching Task (167384ad-7643-43ef-930a-18455f5307a0):
    NAMESPACE                DATA                    TIMESTAMP
    /random/float            0.4507584997241038      2017-06-02 09:47:48.338103527 +0200 CEST
    /random/integer          2.79090919e+08          2017-06-02 09:47:48.338098827 +0200 CEST
    /random/string           Yes definitely          2017-06-02 09:47:48.338109312 +0200 CEST
    Stopping task watch
    ```
    Example `sample-task.json` is given below:
    ```
    {
        "max-failures": 10, 
        "schedule": {
            "interval": "3s", 
            "type": "simple"
        }, 
        "version": 1, 
        "workflow": {
            "collect": {
            "metrics": {
                "/random/*": {}
            }
            }
        }
    }
    ```
    