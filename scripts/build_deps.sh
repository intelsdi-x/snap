### "go got" all dependencies required for building

# exit immediately
set -e

echo "installing build & testing dependencies"
go env

echo
go get -v github.com/golang/lint/golint
go get -v github.com/axw/gocov/gocov
go get -v github.com/mattn/goveralls
go get -v github.com/smartystreets/assertions
go get -v github.com/smartystreets/goconvey
go get -v github.com/smartystreets/goconvey/convey
go get -v github.com/tools/godep
go get -v golang.org/x/tools/cmd/goimports

# cover will be in standard lib from 1.5 for 1.4 we have to live with such kind of hack
# this hack probably  was required for 1.3
#go get -v golang.org/x/tools/cmd/cover
# because separted packages from tools are put in go tooldir this is required
if ! go get code.google.com/p/go.tools/cmd/vet; then go get golang.org/x/tools/cmd/vet; fi
if ! go get code.google.com/p/go.tools/cmd/cover; then go get golang.org/x/tools/cmd/cover; fi

### LIBcontainer
#cd ../../docker/libcontainer; git checkout tags/v1.4.0; cd - going to Godeps
#go get github.com/docker/libcontainer
