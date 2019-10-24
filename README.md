# M5GOperator

A Kubernetes Operator for automatic Mosaic5G deployment on top of a Kubernetes Cluster

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