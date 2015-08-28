# Pulse Examples

This folder contains examples of using Pulse. Each folder contains a README.md which explains the examples.

### Example list

* [influxdb-grafana](influxdb-grafana)
* [tasks](influxdb-grafana)

### Linux pre-setup

Use "linux_prep.sh" to create docker-machine container required by examples.
If you are behind proxy server uncomment and set "ENV http_proxy ..." and "ENV https_proxy ..."
in Dockerfile.

Docker file comes from: https://docs.docker.com/examples/running_ssh_service/
Tested on: Ubuntu 14.04.
Depends on: docker, docker-machine.
