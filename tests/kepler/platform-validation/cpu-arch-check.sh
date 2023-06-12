#!/usr/bin/env bash
#

set -e

echo -e "\n=== Test case 1 (CPU architecture check) Start ===\n"

echo -e "\n--- Start port forwarding ---\n"
kubectl port-forward --namespace=kepler service/kepler-exporter 9102:9102 &

sleep 10

echo -e "\n--- Retrieve 'cpu_architecture' label from Kepler exported metrics ---\n"

metric=$(curl -s http://127.0.0.1:9102/metrics |grep cpu_architecture | grep nodeInfo| cut -d '=' -f 2 | cut -d '"' -f 2)

echo -e "\nmetric label: "$metric

if [[ -z $metric ]];
then
    echo "Test fail! Did not retrieve 'cpu_architecture' label, please check what's the issue!"
    exit 1
fi

echo -e "\n--- Check 'cpu_architeture' on host ---\n"

docker pull gar-registry.caas.intel.com/cpio/platform-validation:2023WW24
docker run --rm -v $(pwd):/output gar-registry.caas.intel.com/cpio/platform-validation:2023WW24
uarch=$(cat cpu-arch)
echo -e "\ncpu architecture: "$uarch
rm -f cpu-arch

echo -e "\n--- Test Result ---\n"

#Compare with local host hardware info check result
if [[ $uarch == $metric ]]
then
    echo "The 'cpu_architecture' label in Kepler metrics is expected, test pass!"
else
    echo "The 'cpu_architecture' label in Kepler metrics is NOT expected, test fail!"
fi

echo -e "\n--- Stop port forwarding ---\n"

# kill "kubectl port-forward xxx" process for kepler-exporter service
pid=$(pgrep kubectl)
if [[ ! -z $pid ]];
then
    kill -9 ${pid}
fi

echo -e "\n=== Test case 1 (CPU architecture check) End ===\n"
