# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.26.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-diki

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION
RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# base
FROM gcr.io/distroless/static-debian13:nonroot AS base
WORKDIR /

############# gardener-extension-diki
FROM base AS diki

COPY --from=builder /go/bin/gardener-extension-diki /gardener-extension-diki
ENTRYPOINT ["/gardener-extension-diki"]
