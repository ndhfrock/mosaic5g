# M5GOperator

A Kubernetes Operator for automatic Mosaic5G deployment on top of a Kubernetes Cluster

It will deploy OpenAirInterface, FlexRAN, Elasticsearch, and Kibana and auto configure everything.

Just a little line command to deploy everything.

## Project layout


| File/Folders   | Purpose                           |
| :---           | :--- |
| cmd       | Contains `manager/main.go` which is the main program of the operator. This instantiates a new manager which registers all custom resource definitions under `pkg/apis/...` and starts all controllers under `pkg/controllers/...`  . |
| pkg/apis | Contains the directory tree that defines the APIs of the Custom Resource Definitions(CRD). Users are expected to edit the `pkg/apis/<group>/<version>/<kind>_types.go` files to define the API for each resource type and import these packages in their controllers to watch for these resource types.|
| pkg/controller | This pkg contains the controller implementations. Users are expected to edit the `pkg/controller/<kind>/<kind>_controller.go` to define the controller's reconcile logic for handling a resource type of the specified `kind`. |
| build | Contains the `Dockerfile` and build scripts used to build the operator. |
| deploy | Contains various YAML manifests for registering CRDs, setting up [RBAC][RBAC], and deploying the operator as a Deployment.
| go.mod go.sum | The [Go mod][go_mod] manifests that describe the external dependencies of this operator. |
| vendor | The golang [vendor][Vendor] directory that contains local copies of external dependencies that satisfy Go imports in this project. [Go modules][go_mod] manages the vendor directory directly. This directory will not exist unless the project is initialized with the `--vendor` flag, or `go mod vendor` is run in the project root. |

## Shell file

- createk8s.sh : shell file to create a kubernetes cluster
- m5goperator.sh : shell file to run this operator on kubernetes
- api.sh : shell file to apply the custom resource, to deploy oai on kubernetes (deploy, update, delete)

## Tutorial
https://hackmd.io/3F62V1JXSqmarN5XLRgoKg?view

Use createk8s.sh

- Edit createk8s.sh first
```shell
KUBE_VERSION="1.15.1-00"
export KUBECONFIG=/home/nadhif/.kube/config  # change nadhif to your hostname
export OPERATOR_NAME=m5g-operator
export MYNAME=${USER}
export MYDNS="140.118.31.99"  # change this to your dns
```

- Build Kubernetes Environtment

```shell=
$ ./createk8s.sh install_req (install kubelet, kubectl, kubeadm)
$ ./createk8s.sh start flannel (start kubernetes flannel) (reccomended)
#or
$ ./createk8s.sh start calico (start kubernetes calico)
```

- Do this if you are using Ubuntu 16.04
```shell=
# Edit coredns
$ kubectl edit cm coredns -n kube-system

# delete ‘loop’ ,save and exit

$ kubectl -n kube-system delete pod -l k8s-app=[kube-dns|core-dns]
# it's fine if the last one doesn't work
```

- If you got this error when creating the kubernetes cluster
```shell=
[kubelet-check] The HTTP call equal to 'curl -sSL http://localhost:10248/healthz' failed with error: Get http://localhost:10248/healthz: dial tcp 127.0.0.1:10248: connect: connection refused.
[kubelet-check] It seems like the kubelet isn't running or healthy.
[kubelet-check] The HTTP call equal to 'curl -sSL http://localhost:10248/healthz' failed with error: Get http://localhost:10248/healthz: dial tcp 127.0.0.1:10248: connect: connection refused.

Unfortunately, an error has occurred:
            timed out waiting for the condition

This error is likely caused by:
            - The kubelet is not running
            - The kubelet is unhealthy due to a misconfiguration of the node in some way (required cgroups disabled)
            - No internet connection is available so the kubelet cannot pull or find the following control plane images:
                - k8s.gcr.io/kube-apiserver-amd64:v1.11.2
                - k8s.gcr.io/kube-controller-manager-amd64:v1.11.2
                - k8s.gcr.io/kube-scheduler-amd64:v1.11.2
                - k8s.gcr.io/etcd-amd64:3.2.18
                - You can check or miligate this in beforehand with "kubeadm config images pull" to make sure the images
                  are downloaded locally and cached.
```

- There may be something wrong with kubelet, so do this
```shell=
$ sudo swapoff -a
$ sudo sed -i '/ swap / s/^/#/' /etc/fstab
```
Then reboot the machine


## M5G Operator Development
[My M5g Operator Development Note](https://hackmd.io/erL2Vn_VRmClrvfymGTlfA?view)


## Setup M5G Operator

```shell=
# apply M5G Operator crd to kubernetes
$ ./m5goperator.sh init 

# start m5g as a container/pod in kubernetes
$ ./m5goperator.sh container start 

# monitor all running pods and its ip address
$ ./m5goperator watch_pods
```

## Deploy OAI Container/Pod
```shell=
# deploy cluster role for api access
$ ./api.sh init

# deploy OAI Container/Pod
$ ./api.sh apply_cr #to use snap ran
$ ./api.sh apply_cr_slicing # to use Samuel's eNB (https://gitlab.com/changshengliusamuel/LTE_Mac_scheduler_with_network_slicing.git -b flexran_LTE_slicing_integration)
```


OAI-ENB will run by itself, wait until your usrp is running and try to connect a ue
