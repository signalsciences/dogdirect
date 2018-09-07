#!/bin/sh
set -ex

go build .
go test .
