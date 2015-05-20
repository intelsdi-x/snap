FROM golang:latest  
ENV GOPATH=$GOPATH:/app
ENV PULSE_PATH=/go/src/github.com/intelsdi-x/pulse/build
RUN apt-get update
RUN apt-get -y install facter
WORKDIR /go/src/github.com/intelsdi-x/
RUN git clone https://<GIT_TOKEN>@github.com/intelsdi-x/gomit.git
WORKDIR /go/src/github.com/intelsdi-x/pulse
ADD . /go/src/github.com/intelsdi-x/pulse
RUN go get github.com/tools/godep
RUN go get golang.org/x/tools/cmd/goimports
RUN go get golang.org/x/tools/cmd/vet
RUN go get golang.org/x/tools/cmd/cover
RUN scripts/deps.sh
RUN make
