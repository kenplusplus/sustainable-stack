# Intel Platform Validation Guide in Kepler

## Quick Start

### Platform prerequisite setup

#### Tools and packages installation

**CPUID**

Intel 4th Gen Xeon Scalable Proecessor (Codename: `Sapphire Rapids`, aka. `SPR`) was first released at 2023/1/10.

`cpuid` is the most popular tool on Linux distros to check the CPU detailed characteristics.

We have confirmed that `cpuid` does not fully support `SPR` until version `20230505`.

On `Ubuntu` Linux distro, the latest version of cpuid is still `20211210`, so we currently prefer `RHEL` Linux distro to do validation on `SPR` and subsequent platforms(`EMR`/`GNR`/`SFT`, etc)

For `RHEL`, please use below commands to check cpuid version and upgrade if necessary:
```
$ cpuid --version
cpuid version 20230120      ==>earlier than 20230505, need upgrade to support SPR!

$ sudo dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm

$ sudo dnf install -y cpuid

$ cpuid --version
cpuid version 20230505
```

**Others**

Other Requirements come from Kepler project, please check [here](https://sustainable-computing.io/installation/kepler/).

Typically, your test machine should have installed:
* Kernel 4.18+
* Access to a Kubernetes cluster
* kubectl v1.21.0+
* go v1.18+


#### Confirm if $USER has joined `wheel`/`sudo` and `docker` group

For `RHEL` system, $USER who wants to execute `sudo` command, need to join `wheel` group;

For `Ubuntu` system, $USER who wants to execute `sudo` command, need to join `sudo` group;

$USER who wants to execute `docker` command without `sudo`, need to join `docker` group also.

```
$ echo $USER
jie
$ groups
jie wheel docker
```

#### Configure proxy (Optional)

For Intel internal network, we need to configure proxy properly, general proxy could be configured at `~/.bashrc` with Environment Variables such as:
`http_proxy`, `HTTP_PROXY`, `https_proxy`, `HTTPS_PROXY`, `no_proxy` and `NO_PROXY`

We need to add extra FQDN("kind-registy") into $no_proxy/$NO_PROXY due to [Kind](https://kind.sigs.k8s.io/) limitation:

```
export no_proxy=$no_proxy,kind-registry
export NO_PROXY=$NO_PROXY,kind-registry
```

We also need to add Docker Client engine proxy at `~/.docker/config.json`, follow Docker official [guide](https://docs.docker.com/network/proxy/#configure-the-docker-client), change specific URLs as your local configuration.

### Deploy Kepler in local Kubernetes cluster

Kepler uses [Kind](https://kind.sigs.k8s.io/) as the local Kubernets cluster provider, and supports three deployment scenarios:

* Deploy from [kepler-helm-chart](https://github.com/sustainable-computing-io/kepler-helm-chart/blob/main/README.md)
* Deploy from [kepler-operator](https://github.com/sustainable-computing-io/kepler-operator/blob/v1alpha1/README.md)
* Deploy from [source code](https://sustainable-computing.io/installation/kepler/#deploy-from-source-code)

Here take the `Deploy from source code` as example to demonstrate the most flexibility.

#### Deploy from source code

**Step 1: Clone code.**

* Clone latest code:
```
git clone https://github.com/sustainable-computing-io/kepler.git
```
or

* Clone code with release tag. i.e. v0.5 release:
```
git clone -b v0.5 https://github.com/sustainable-computing-io/kepler.git --depth=1
```

**Step 2: Bring up Kind cluster.**

```
cd kepler
make cluster-up
```

After this step, you could see the cluster status like this:
```
$ kubectl get pods -A
NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
kube-system          coredns-565d847f94-ggvr8                     1/1     Running   0          16m
kube-system          coredns-565d847f94-vt8kn                     1/1     Running   0          16m
kube-system          etcd-kind-control-plane                      1/1     Running   0          16m
kube-system          kindnet-lxqtq                                1/1     Running   0          16m
kube-system          kube-apiserver-kind-control-plane            1/1     Running   0          16m
kube-system          kube-controller-manager-kind-control-plane   1/1     Running   0          16m
kube-system          kube-proxy-256b4                             1/1     Running   0          16m
kube-system          kube-scheduler-kind-control-plane            1/1     Running   0          16m
local-path-storage   local-path-provisioner-684f458cdd-js8hx      1/1     Running   0          16m
monitoring           prometheus-k8s-0                             2/2     Running   0          15m
monitoring           prometheus-operator-7b64d465b9-7v6b8         2/2     Running   0          15m
```

**Step 3: Deploy Kepler.**

```
make cluster-sync
```

After this step, you could see `kepler-exporter` pod running in the cluster:

```
$ kubectl get pods -A
NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
kepler               kepler-exporter-bj99g                        1/1     Running   0          38s
kube-system          coredns-565d847f94-ggvr8                     1/1     Running   0          19m
kube-system          coredns-565d847f94-vt8kn                     1/1     Running   0          19m
kube-system          etcd-kind-control-plane                      1/1     Running   0          19m
kube-system          kindnet-lxqtq                                1/1     Running   0          19m
kube-system          kube-apiserver-kind-control-plane            1/1     Running   0          19m
kube-system          kube-controller-manager-kind-control-plane   1/1     Running   0          19m
kube-system          kube-proxy-256b4                             1/1     Running   0          19m
kube-system          kube-scheduler-kind-control-plane            1/1     Running   0          19m
local-path-storage   local-path-provisioner-684f458cdd-js8hx      1/1     Running   0          19m
monitoring           prometheus-k8s-0                             2/2     Running   0          18m
monitoring           prometheus-operator-7b64d465b9-7v6b8         2/2     Running   0          18m
```

You could see the `kepler-exporter` service running like this:

```
$ kubectl get service -n kepler
NAME              TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
kepler-exporter   ClusterIP   None         <none>        9102/TCP   40s
```

### Retrieve `cpu_architecture` label from Kepler exported metrics

Use port forwarding mechanism to forward Kepler metrics from internal default `9102` port to `localhost` port:

```
$ kubectl port-forward --namespace=kepler service/kepler-exporter :9102
Forwarding from 127.0.0.1:44305 -> 9102
Forwarding from [::1]:44305 -> 9102
```

Retrieve `cpu_architecture` specific metric:

```
$ curl -s http://127.0.0.1:44305/metrics |grep cpu_architecture | grep nodeInfo
kepler_node_nodeInfo{cpu_architecture="Sapphire Rapids"} 1
```

### Compare with local host hardware info check result

For `RHEL` Linux distro, users could directly call `cpuid` command to check the host cpu architecture.

```
$ cpuid -1 |grep uarch
   (uarch synth) = Intel Sapphire Rapids {Golden Cove}, Intel 7
```

To make the validation host OS agnostic, we leverage `Docker` technology to provide pre-built Docker image for users to check the host cpu info properly:

```
$ docker pull gar-registry.caas.intel.com/cpio/platform-validation:2023WW24

$ docker run -it --rm -v $(pwd):/output gar-registry.caas.intel.com/cpio/platform-validation:2023WW24

$ cat cpu-arch
Sapphire Rapids
```

The `Dockerfile` for such Docker image could be found under `build` directoy.

The `platform-validation.sh` is under the same diretory, which is built as `ENTRYPOINT` of the Docker image.

They are kept up-to-date and properly tagged (Currently we use the release work week as tag, i.e. 2023WW24).

The tag may change in below scenarios:

* New test cases are integrated, the entrypoint script should update.
* New CPUs or platforms are published, cpuid or other tools' version may upgrade.
* cpuid command's output format changes.

### Automation

All the above validation steps have been automated, the scripts could be found under `tests/kepler/platform-validation` directory.

`setup.sh` includes case setup phase operations.

`cleanup.sh` includes case teardown phase operations.

Both of them could be shared by all the test cases. 

`cpu-arch-check.sh` includes the test jobs of `cpu_architecure check` case.

Currently, for this test case, the execution steps are as follows:
```
$ ./setup.sh
$ ./cpu-arch-check.sh
$ ./cleanup.sh
```

## Further validation test cases

TBD
