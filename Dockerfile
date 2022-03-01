# SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

# Build
FROM onosproject/golang-build:latest as build

ENV CGO_ENABLED=0

COPY ./ $GOPATH/src/github.com/onosproject/subscriber-dns
WORKDIR $GOPATH/src/github.com/onosproject/subscriber-dns

RUN go build -o /go/bin/subdns ./cmd/subdns

# Deploy
FROM alpine:latest
RUN apk add bash openssl curl libc6-compat

WORKDIR /home/subscriber-dns

COPY --from=build /go/bin/subdns /usr/local/bin/