# Train Data Collector

This collector provides a framework for fully automated collection of various events of Intel accelerator including AMX, QAT, DSA, etc, hardware counters and RAPL.

The main purpose of this collector is to collect the input data needed to train the energy consumption model (Please view modeling directory for more detailed information). This collector only cares about the accurate collection of raw data, the model training is not in the scope of this collector.

## AMX

[Cloud Native AI Pipeline](https://github.com/intel-innersource/os.linux.cloud.mvp.elastic-inference/tree/main) is the reference AI workload to use AMX. Below operations all depend on it, please deploy it as a prerequisite.

### 0. (Optional) Check if the AI Pipeline is deployed successfully

```bash
kubectl logs -f <amx infer pod>
2023-06-08 08:13:09,972 INFO   [inferbase] [pose-bf16-amx] Drop frame speed: 39.360942 FPS
2023-06-08 08:13:20,068 INFO   [inferbase] [pose-bf16-amx] Infer speed: 10.404967 FPS
2023-06-08 08:13:20,069 INFO   [inferbase] [pose-bf16-amx] Drop frame speed: 40.331636 FPS
2023-06-08 08:13:30,168 INFO   [inferbase] [pose-bf16-amx] Infer speed: 10.299783 FPS
```

### 1. Build AMX collector

```bash
cd collector
make amx
```

### 2. Start AMX collector

- Use `help` to list all supported parameter

```bash
./amx_collector --help
  -d, --deployName string   The deployment that needs to be scaled
  -e, --events strings      PMU events that need to be monitored
  -f, --freq int            CPU frequency when running AMX (default 2400000)
      --kubeconfig string   (optional) absolute path to the kubeconfig file (default "/path/to/.kube/config")
  -n, --namespace string    Namespace where the scaling deployment in (default "default")
  -p, --pids ints           PIDs that need to be monitored
      --pods int            Number of AMX running pods (default -1)
```

- Run the AMX collector

```bash
sudo ./amx_collector -e <events> -f <target frequency> -p <amx pids> --kubeconfig /path/to/.kube/config -d <deployment> -n <namespace>

# Example
sudo ./amx_collector -e exe.amx_busy,cycles,instructions,cache-misses -f 800000 -p 1693337 --kubeconfig /home/lei/.kube/config -d ei-infer-pose-bf16-amx-deployment
```

### 3. Results

All collected results will be stored in the csv files.

- The AMX related events are stored in a file with `amx_event` prefix.

```bash
[lei@cpio-sprqct-prc5 cmd]$ cat amx_event_20230606072833.csv
exe.amx_busy,cycles,instructions,cache-misses,cpu_freq,inferpod_num
5577603904,133420919664,222030223148,327704261,2400000,1
5608369173,133251668274,223961179044,263582724,2400000,1
5581481684,133125208309,222971830007,234859502,2400000,1
5584137060,133492974164,222666484187,275202819,2400000,1
5579176895,133228043835,223406543027,287462679,2400000,1
```

- The RAPL values are stored in a file with `energy_result` prefix.

```bash
[lei@cpio-sprqct-prc5 cmd]$ cat energy_result_20230608213537.csv
Package,Memory
21090211573,289593489
21143349865,290143646
21043136560,288950981
21050706546,288780233
```

## QAT

### 0. Prerequisites

Hardware Requirements

* [Intel® 4XXX Series](https://www.intel.com/content/www/us/en/products/details/processors/xeon/scalable.html)

Software Requirements

* [Intel® QuickAssist Technology Drive for linux - HW Version 2.0](https://www.intel.com/content/www/us/en/download/765501/intel-quickassist-technology-driver-for-linux-hw-version-2-0.html)

* [QATzip](https://github.com/intel/QATzip)

### 1. Build QAT collector


```bash
cd collector
make qat
```

### 2. Start QAT collector

* Use `help` to list all supported parameter

```bash
./qat_collector --help
Usage of ./qat_collector:
  -f, --freq string            CPU frequency when running QAT (default "2400000")
  -i, --inputDirPath string    Input Dir Path For QATzip
  -o, --outputDirPath string   Output Dir Path For QATzip
  -r, --resultDirPath string   Result Dir Path For QATzip
```

* Run the collector

```bash
sudo ./qat_collector -f <target frequency> -i <input directory> -o <output directory> -r <result directory>

# Example
sudo ./qat_collector -f 800000 -i ../../test/input -o ../../test/output -r ../../test/result
```

### 3. Results

All collected results will be stored in a csv file.

```bash
cat result.csv
# Basic data: cpu_freq
# Telemetry values: time_cnt_sum, pci_trans_sum, latency_sum, bw_in_sum, bw_out_sum, cpr_sum, dcpr_sum
# RAPL values: pkg_energy, dram_energy
cpu_freq,time_cnt_sum,pci_trans_sum,latency_sum,bw_in_sum,bw_out_sum,cpr_sum,dcpr_sum,pkg_energy,dram_energy
2400000,25,0,2245690,100855,23007,219,79.31,924294.84,4051.09
3800000,1,0,133028,8454,65,27,5,86433.1343,750.432278
```

And qzip log will be stored in the qzip.log.

```bash
cat qzip.log
Reading input file ../../test/input/test1 (3518119 Bytes)
Compressing...
Time taken:        4.533 ms
Throughput:     6208.902 Mbit/s
Space Savings:     0.212 %
Compression ratio: 1.002 : 1
Reading input file ../../test/input/test2.gz (1789071 Bytes)
Decompressing...
Time taken:        3.728 ms
Throughput:     3837.536 Mbit/s
```

## DSA
TBD