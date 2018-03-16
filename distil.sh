#!/usr/bin/env bash
set -eux

MOON="moonshine"
DIR="release"
mkdir "$DIR"
env GOOS=windows GOARCH=386 go build
mv "$MOON".exe "${DIR}/${MOON}"-win386.exe
env GOOS=windows GOARCH=amd64 go build
mv "$MOON".exe "${DIR}/${MOON}"-win64.exe
env GOOS=linux GOARCH=amd64 go build
mv "$MOON" "${DIR}/${MOON}"-linux64
env GOOS=darwin GOARCH=386 go build
mv "$MOON" "${DIR}/${MOON}"-darwin386
env GOOS=darwin GOARCH=amd64 go build
mv "$MOON" "${DIR}/${MOON}"-darwinAmd64
