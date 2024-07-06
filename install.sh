#!/bin/sh

BINARY='/usr/local/bin'

echo "Building dkr"
go build dkr.go

echo "Installing dkr to $BINARY"
install -v dkr $BINARY

echo "Removing the build"
rm dkr
