FROM golang:latest

ENV GOPATH=$GOPATH:/go
COPY ./ /go/src/github.com/intelsdi-x/snap
RUN apt-get update \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/intelsdi-x/gomit.git /go/src/github.com/intelsdi-x/gomit \
    && /go/src/github.com/intelsdi-x/snap/scripts/deps.sh \
    && make -C /go/src/github.com/intelsdi-x/snap
WORKDIR /go/src/github.com/intelsdi-x/snap

