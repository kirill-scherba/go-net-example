# Teonet-go


## Requirements

Teonet-go requires Go 1.2 or newer. Download the latest from: https://code.google.com/p/go/downloads/

Git is required for installing via go get.

A C compiler (gcc or clang) and the openssl library are required too.

On Debian and Ubuntu, install the development packages:

    apt-get install build-essential libssl-dev

On CentOS and RHEL:

    yum install gcc libssl-dev


## Installation

Make sure $GOROOT and $GOPATH are set, and install:

    go get


## Running

To run main teonet developer application:

    cd teonet
    run go .
