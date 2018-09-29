#!/bin/sh
set -ex

go build . ./hostmetrics ./cmd/ddd/...
go test .
