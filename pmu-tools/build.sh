#/usr/bin/bash

set -ex

CURR_DIR=$(dirname "$(readlink -f "$0")")

get_origin() {
    rm -rf pmutools
    echo "Download origin package..."
    git clone https://github.com/andikleen/pmu-tools.git pmutools
}

prepare() {
    echo "Prepare..."
    cp __init__.py pmutools/
    python3 -m pip install --upgrade build
}

build() {
    echo "Build..."
    python3 -m build
}

pushd ${CURR_DIR}
get_origin
prepare
build
popd
