#!/usr/bin/env bash
#

set -e

CLONED=true

usage() {
    cat << EOM
Usage: $(basename "$0") [Parameter]

Paremeter:
    Should be valid kepler repo tag(i.e. v0.5, v0.4.11, etc) or "latest"(main branch).
    If not set, $(basename "$0") will clone "main" branch of kepler.
EOM
}

process_arg() {
    if [ "$#" -gt "1" ];
    then
        echo "Invalid parameter number!"
        usage
	exit 1
    fi
    rm -rf kepler
    if [ $# == 0 ] || [ "$1" == "latest" ]
    then
        echo "Cloning the latest Kepler code..."
        git clone https://github.com/sustainable-computing-io/kepler.git
    else
        git clone -b $1 https://github.com/sustainable-computing-io/kepler.git --depth=1
	if [ $? -eq 0 ]
        then
            echo "Cloned $1 tag of Kepler code"
        else
            echo "Invalid tag, please check your input!"
            CLONED=false
	    exit 1
        fi
    fi
}

main() {
    #bring up kind cluster
    if [ ! $CLONED ]
    then
        exit 1
    fi

    cd kepler && make cluster-up

    #deploy kepler
    make cluster-sync

    # round for 3 times and each for 60s
    # check if the rollout status is running
    deploy_status=1
    for i in 1 2 3
    do
        echo "check deployment status for round $i"
        kubectl rollout status daemonset kepler-exporter -n kepler --timeout 60s
        #check rollout status
        if [ $? -eq 0 ]
        then
            deploy_status=0
            break
        fi
    done

    if test $[deploy_status] -eq 1
    then
        echo "Check the status of the kepler-exporter"
        kubectl -n kepler describe daemonset.apps/kepler-exporter
        echo "Check the logs of the kepler-exporter"
        kubectl -n kepler logs daemonset.apps/kepler-exporter
        exit
    fi

    echo "Kepler has been deployed successfully"
}

process_arg "$@"
main
