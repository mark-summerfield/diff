#!/bin/bash
clc -s -e diff2_test.go
cat Version.dat
go mod tidy
go fmt .
echo no staticcheck
# staticcheck .
go vet .
echo no golangci-lint
# golangci-lint run
git st
