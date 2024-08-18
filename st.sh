#!/bin/bash
clc -s -e diff2_test.go
cat Version.dat
go mod tidy
go fmt .
staticcheck .
go vet .
golangci-lint run
git st
