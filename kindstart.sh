#!/bin/bash

# To add more images, visit: https://hub.docker.com/r/kindest/node
# kubectl cluster-info --context kind-v119


kind delete --name v119 cluster


cat <<EOF | kind create cluster --name v119 --config -
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.19.7@sha256:a70639454e97a4b733f9d9b67e12c01f6b0297449d5b9cbbef87473458e26dca
EOF

