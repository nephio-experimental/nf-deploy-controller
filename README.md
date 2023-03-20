# NFDeploy Controller
NFDeployment controller watches and processes the NFDeployment custom resources. It runs on Nephio's management cluster.

## Description
NFDeployment controller watches and process Nephio's NFDeployment custome resources. It primarily performs the following functions:
1. interprets CRs embedded inside NFDeployment CR
1. fetches corresponding CRs to inject variants
1. tracks and aggregates statuses of each of NF instance's deployment

The following diagram depicts NFDeployment controller's rule in Nephio R1

![NFDeployment Controller's role](./img/nfdeploy.jpg)

1. A NFDeployment CR is applied to the management cluster
1. NFDeployment controller executes the "fan out" process
1. Essentially, the fanout process is cloning source package from catalog to each of the target deployment repo
1. Independently, NFDeployment controller advertises reachability info to the watcher agent running on the workload clusters
1. User manually provides input to the cloned package(s), then proposes and approves the package(s) to be deployed
; the deployed package(s) are read by configsync running on workload cluster, and subsequently applies to workload cluster
1. NF vendor specific NF operator processes Nephio CR, then deploys an instance of NF
1. watcher-agent running on workload cluster reads Nephio CR status, and reports back to edge-watcher on NFDeployment controller


**NFDeployment**

NFDeployment consists of a deployment unit to track, which includes one or more NF instance, where each instance includes:
- ID
- name of cluster
- NF type (AMF, SMF, UPF)
- NF flavor (user defined, ex: small, medium, large)
- NF vendor
- NF vendor's NF software version
- Connectivities (list of neighbor names, i.e., NFDeploy.Spec.Id)

The high level NFDeployment deployment unit also contains PLMN and overall capacity

On a high level, NFDeployment controller consists of two entities, hydration and deployment:

![NFDeployment Controller internal high-level design](./img/nfdeploy-controller-internal.jpg)

The hydration entity examines the NFDeployment CR, clones source package based on vendor and version, and clones those packages to a deployment repo based on cluster's name. It primarily interacts with Porch (now), and in the future will primarily interacts with PackageVariant controller.

The deployment entity takes the topology information from NFDeployment CR, and builds a relationship graph to track each individual NF specific status. As part of this entity, NFDeploy controller creates and maintains an instance of EdgeWatcher object with a newly created gRPC server to collect workload cluster selected CRs statuses. The changes in any individual status is reflected on NFDeployment's own status via this deployment entity.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
make install
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/nfdeploy:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/nfdeploy:tag
```

4. Actually deploy the controller:

```sh
kubectl apply -f config/deployment/deployment.yaml
```

**Note** It is expected that your Kubernetes cluster will be running v1.11.0 of all the deployments running in the **cert-manager** namespace, namely: cert-manager, cert-manager-webhook, and cert-manager-cainjector. You can check via `kubectl describe <pod> -n cert-manager`. In case any of them isn't at v1.11.0, you can update via `kubectl edit deployment.apps/cert-manager{-webhook | -cainjector} -n cert-manager`

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO: Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

**NOTE:** `make run` does **NOT** work as the controller requires some environment variables and mountPaths to operate
