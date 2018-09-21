#!/bin/sh
set -ex

go build . ./hostmetrics
go test .
