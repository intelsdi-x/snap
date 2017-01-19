go get github.com/Masterminds/glide
go get -d github.com/Snap-for-Windows/snap
cd %GOPATH%\src\github.com\Snap-for-Windows\snap
glide install
fo install
cd \cmd\snapctl
go install
