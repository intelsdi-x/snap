FROM golang:latest
ENV GOPATH=$GOPATH:/app
RUN apt-get update
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/snap
ADD . /go/src/github.com/intelsdi-x/snap
RUN scripts/deps.sh
RUN make
