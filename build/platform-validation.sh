#!/usr/bin/env bash
#

set -e

uarch=$(cpuid -1 |grep uarch | cut -d ' ' -f 8,9)

echo $uarch > /output/cpu-arch
