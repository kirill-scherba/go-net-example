# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y wget git gcc libssl-dev

RUN wget -P /tmp https://golang.org/dl/go1.15.5.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf /tmp/go1.15.5.linux-amd64.tar.gz
RUN rm /tmp/go1.15.5.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"


# ENV GIT_SSL_NO_VERIFY=true

# Copy the local package files to the container's workspace.
# ADD . /go/src/github.com/kirill-scherba/teonet-go/teonet/app/teocdb

# Build the teocdb command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
# RUN apt update && apt install git
# RUN go get github.com/kirill-scherba/teonet-go

RUN go get github.com/kirill-scherba/teonet-go/teonet
RUN cd /go/src/github.com/kirill-scherba/teonet-go/teonet/app/teocdb && go get && go install

# Run the outyet command by default when the container starts.
CMD /go/bin/teocdb -p 11011 teo-cdb

# Document that the service listens on port 11011.
EXPOSE 11011/udp

# run example: docker run -d -it --restart=always --name teocdb -p 11011:11011/udp teocdb