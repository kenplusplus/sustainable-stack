#! /bin/bash

EVENT=""
TIME=60
PID=""
OUTPUT="perf.txt"

usage() {
    cat << EOM
Usage: $(basename "$0") [OPTION]...
  -e   The PMU event
  -t   The time of perf stat monitor (Unit: second)
  -p   The process id
  -o   The output file
  -h   Show this
EOM
    exit 0
}

process_args() {
    while getopts ":e:t:p:o:h" option; do
        case "${option}" in
            e) EVENT=${OPTARG};;
            t) TIME=${OPTARG};;
            p) PID=${OPTARG};;
            o) OUTPUT=${OPTARG};;
            h) usage;;
        esac
    done

    if [[ -z ${EVENT} ]]; then
        usage
        echo "Must set PMU event"
        exit 1
    fi
}

run_perf() {
    if [[ -z ${PID} ]]; then
        perf stat -e ${EVENT} -o ${OUTPUT} sleep ${TIME}
    else
        perf stat -e ${EVENT} -p ${PID} -o ${OUTPUT} sleep ${TIME}
    fi
}

process_args $@
run_perf