# Teonet-go

# About

Teonet-go is golang implementation of Teonet C library. It is communication library to process network / cloud transport between microservices. Teonet-go uses UDP to communicate between its network peers. There UDP packets are encrypted with unique keys. Teonet-go uses own UDP based protocol called TR-UDP for real time communications that allows sending messages with low latency and provides protocol reliability features.

***This project is under constraction. The documentation and examples are not finesed yet. Look at Teonet C Library to use it in production.

## Requirements

Teonet-go requires Go 1.2 or newer. Download the latest from: <https://code.google.com/p/go/downloads/>

Git is required for installing via `go get`.

A C compiler (gcc or clang) and the openssl library are required too.

On Debian and Ubuntu, install the development packages:

    apt-get install build-essential libssl-dev

On CentOS and RHEL:

    yum install gcc libssl-dev

## Installation

Make sure $GOROOT and $GOPATH are set, and install:

    go get github.com/kirill-scherba/teonet-go

## Running

To run main teonet developer application:

    cd teonet
    go run .
