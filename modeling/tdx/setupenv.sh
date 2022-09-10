#!/bin/bash

export SCRIPT_DIR=$(dirname $(readlink -f '$0'))
export PYTHONPATH=${SCRIPT_DIR}/../../external/tdx-tools/utils/pycloudstack