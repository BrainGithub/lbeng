#!/bin/bash
tar -xvf vendor.go > /dev/null
docker run --rm --name build-dp_mgt_$(basename $PWD) -v "$PWD":/go/src/lbeng -w /go/src/lbeng -e GOOS=linux -e GOARCH=amd64 dr.z/golang_ub:latest go build -ldflags "-s -w" -o lbeng ./main.go
rm vendor -rf
#find $PWD -not -name 'lbeng conf' -delete
#shopt -s extglob
#rm -rf !(conf|lbeng)
