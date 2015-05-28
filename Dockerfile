FROM golang:1.4.2  

#PROXYPLACEHOLDER

RUN apt-get update && \
  apt-get -y install facter 

# we use this GOPATH because set we have to /go/bin gratis
# this path would be used for install all binaries required to build like godep
RUN mkdir -p /go && mkdir -p /scripts
ENV GOPATH=/go
# prepare or building deps before add all working directory
# speedups rebuilding because docker cache is valid unless build_deps.sh or Makefile is changed
ADD scripts/build_deps.sh scripts/build_deps.sh
RUN ./scripts/build_deps.sh

ENV PULSE_PATH=/go/src/github.com/intelsdi-x/pulse/build
ADD . /go/src/github.com/intelsdi-x/pulse
WORKDIR /go/src/github.com/intelsdi-x/pulse
# this would build and test everything
CMD make test
