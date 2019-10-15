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
# Run:
#
#  docker run --rm -it teonet-go go run . teo-dteo
#
# Publish to github:
#
#  docker login docker.pkg.github.com -u USERNAME -p TOKEN
#  docker tag teonet-go docker.pkg.github.com/kirill-scherba/teonet-go/teonet-go:0.5.0
#  docker push docker.pkg.github.com/kirill-scherba/teonet-go/teonet-go:0.5.0
#
FROM golang:1.13.1

WORKDIR /go/src/github.com/kirill-scherba/teonet-go/teonet
RUN apt-get update; apt-get install -y libssl-dev
COPY . ../

RUN go get

CMD ["go", "run ."]
