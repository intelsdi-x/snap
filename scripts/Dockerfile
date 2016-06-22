FROM golang:latest  
ENV GOPATH=$GOPATH:/app
ENV SNAP_PATH=/go/src/github.com/intelsdi-x/snap/build
RUN apt-get update && \
    apt-get -y install facter
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/snap
ADD . /go/src/github.com/intelsdi-x/snap
RUN go get golang.org/x/tools/cmd/goimports && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/smartystreets/goconvey
RUN make
