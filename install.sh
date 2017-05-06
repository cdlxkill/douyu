#!/bin/bash
if [ ! -f install.sh ]; then
echo install must be run within its container folder 1>&2
exit 1
fi

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

gofmt -w src
GOOS=linux GOARCH=amd64 go install douyu
GOOS=windows GOARCH=amd64 go install douyu
GOOS=windows GOARCH=386 go install douyu
export GOPATH="$OLDGOPATH"

echo finished
