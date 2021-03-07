#!/bin/bash

unset GOSUMDB
unset GOPROXY

make generate
make manifest

go fmt ./...
make docker-build IMG=memcached-operator:v0.0.1

kind --name v118 load docker-image memcached-operator:v0.0.1
kind --name v119 load docker-image memcached-operator:v0.0.1

# Remove old CRD
make uninstall
make install


make undeploy IMG=memcached-operator:v0.0.1
make deploy IMG=memcached-operator:v0.0.1
