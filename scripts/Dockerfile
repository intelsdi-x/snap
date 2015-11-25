FROM golang:latest  
ENV GOPATH=$GOPATH:/app
ENV SNAP_PATH=/go/src/github.com/intelsdi-x/snap/build
RUN apt-get update && \
    apt-get -y install facter
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://<GIT_TOKEN>@github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/snap
ADD . /go/src/github.com/intelsdi-x/snap
RUN go get github.com/tools/godep && \
    go get golang.org/x/tools/cmd/goimports && \
    go get golang.org/x/tools/cmd/vet && \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/smartystreets/goconvey
RUN scripts/deps.sh
RUN make
