#!/bin/bash

set -e

# format code
gofmt -l -w -s .

# something is wrong with my go structure.
gometalinter --deadline 30s ./... | grep -v "/usr/local/go/src"

# attempt build
cd ./cmd/integrity
go build
go clean