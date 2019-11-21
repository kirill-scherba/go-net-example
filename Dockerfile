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

# Docker builder
# 
FROM golang:1.13.4 AS builder

WORKDIR /go/src/github.com/kirill-scherba/teonet-go/teonet
RUN apt update && apt install -y libssl-dev
COPY . ../

RUN go get && go install

CMD ["teonet"]

# #############################################################
# compose production image
FROM ubuntu:19.04 AS production
WORKDIR /app

# runtime dependencies
RUN apt update && apt install -y libssl1.1
# libssl1.1  -- 109 MB
# openssl    -- 110 MB
# libssl-dev -- 117 MB

# install previously built application
COPY --from=builder /go/bin/* /usr/local/bin/
CMD ["/usr/local/bin/teonet"]  
