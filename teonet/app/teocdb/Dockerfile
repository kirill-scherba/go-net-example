# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.15-alpine


# ENV GIT_SSL_NO_VERIFY=true

# Copy the local package files to the container's workspace.
# ADD . /go/src/github.com/kirill-scherba/teonet-go/teonet/app/teocdb

# Build the teocdb command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
# RUN apt update && apt install git
# RUN go get github.com/kirill-scherba/teonet-go

RUN cd /go/src/github.com/kirill-scherba/teonet-go/teonet/app/teocdb && go get && go install

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/teocdb

# Document that the service listens on port 8080.
EXPOSE 11011/udp