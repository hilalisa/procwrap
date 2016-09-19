### Build info

make build.sh executable

E.G. to build on golang:1.7.1-alpine image execute the following from the root of the repo:

./build/build.sh `pwd` `pwd`/build/output golang:1.7.1-alpine amd64_alpine github.com/m3europe/procwrap procwrap

This depends on the GOPATH being /usr/local/go in the container and will create a binary at ./build/output/amd64_alpine/procwrap


