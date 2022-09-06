# PMU Tools

A python package for https://github.com/andikleen/pmu-tools

## Build

```
$ ./build.sh
```

The script will pull pmu-tools and build wheel package.  

## Install

```
$ sudo su
$ python3 -m pip install dist/pmutools-*.whl 
```

pmu-tools needs to be run as root.
Note that if the installation shows below warning, please add '/usr/local/bin' to you PATH environment.

`WARNING: The script toplev is installed in '/usr/local/bin' which is not on PATH.`

## Run

```
toplev --help
```