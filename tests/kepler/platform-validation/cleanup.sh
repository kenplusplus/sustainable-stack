#!/usr/bin/env bash
#

set -e

# uninstall kepler
cd kepler && make cluster-clean

# cleanup kind cluster & kind local registry

#TODO: local-dev-cluster v0.0.0 doest not support ./main.sh down.
# cd local-dev-cluster && ./main.sh down

./local-dev-cluster/kind/.kind delete cluster --name=kind
docker rm -f kind-registry

# cleanup kepler source code
cd ../ && rm -rf kepler
