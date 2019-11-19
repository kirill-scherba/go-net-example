# Copyright 2019 Teonet-go authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
#
# Teonet-go docker file
#
# Build:
#
#  docker build -t teonet-go .
#
# Publish to github:
#
#  docker login docker.pkg.github.com -u USERNAME -p TOKEN
#  docker tag teonet-go docker.pkg.github.com/kirill-scherba/teonet-go/teonet-go:0.5.0
#  docker push docker.pkg.github.com/kirill-scherba/teonet-go/teonet-go:0.5.0
#
# Publish to local repository:
#
#  docker tag teonet-go 192.168.106.5:5000/teonet-go
#  docker push 192.168.106.5:5000/teonet-go
#
# Run docker container:
#
#  docker run --rm -it teonet-go go run . teo-dteo
#
# Run in swarm claster:
#
#  docker service create --constraint 'node.hostname == teonet' --network teo-overlay --name teonet-go -t 192.168.106.5:5000/teonet-go teonet -a 5.63.158.100 -r 9010 -n teonet teo-go-01
#
#
FROM golang:1.13.4

WORKDIR /go/src/github.com/kirill-scherba/teonet-go/teonet
RUN apt update; apt install -y libssl-dev
COPY . ../

RUN go get && go install

#CMD ["go", "run", "."]
CMD ["teonet"]

# # docker build --tag=teobot -f./Dockerfile.ubuntu19 ./
# # #############################################################
# # populate build environment
# FROM ubuntu:19.04 AS builder
# WORKDIR /app

# # Autoconf dependencies
# RUN apt update
# RUN apt install -y libncurses5-dev libncursesw5-dev libreadline-dev libcrypto++-dev libcurlpp-dev libglib2.0-dev libev-dev libcurl4-openssl-dev libssl-dev
# RUN apt install -y automake intltool libtool m4 doxygen
# RUN apt install -y libcunit1-dev libcppunit-dev cpputest

# # copy sources
# COPY . .
# # build all programs
# RUN ./compact.sh


# # #############################################################
# # compose production image
# FROM ubuntu:19.04 AS production
# WORKDIR /app

# # runtime dependencies
# # RUN apt install libncurses5 libncursesw5 libreadline8 libcrypto++6 libcurlpp0
# RUN apt update && apt install -y libreadline8 libcrypto++6 libcurlpp0

# # install previously built artifacts
# COPY --from=builder /app/.pkgroot/usr/local/bin/* /usr/local/bin/
# CMD ["/usr/local/bin/teogsint"]  