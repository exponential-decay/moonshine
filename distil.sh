#!/usr/bin/env bash
set -eux

MOON="moonshine"
DIR="release"
mkdir "$DIR"
export GOOS=windows
export GOARCH=386
go build
mv "$MOON".exe "${DIR}/${MOON}"-win386.exe
export GOOS=windows
export GOARCH=amd64
go build
mv "$MOON".exe "${DIR}/${MOON}"-win64.exe
export GOOS=linux
export GOARCH=amd64
go build
mv "$MOON" "${DIR}/${MOON}"-linux64
export GOOS=darwin
export GOARCH=386
go build
mv "$MOON" "${DIR}/${MOON}"-darwin386
export GOOS=darwin
export GOARCH=amd64
go build
mv "$MOON" "${DIR}/${MOON}"-darwinAmd64
export GOOS=
export GOARCH=
