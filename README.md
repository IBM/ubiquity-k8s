# Ubiquity-k8s
This projects contains the needed components to manage persistent storage for kubernetes through [Ubiquity](https://github.com/IBM/ubiquity) service.
The repository contains mainly two components that could be used separately or combined according to the requirements:
- Ubiquity dynamic provisioner
- Ubiquity flex driver

# Ubiquity dynamic provisioner for k8s

Ubiquity provisioner facilitates creation and deletion of persistent storage, via [ubiquity](https://github.com/IBM/ubiquity) service, within kubernetes

### Prerequesites
* Functional [kubernetes]() environment (v1.5.0 or higher is required for flexvolume support)
* Spectrum-Scale client must be installed on the nodes
* [Ubiquity](https://github.com/IBM/ubiquity) service must be running
* Install [golang](https://golang.org/) and setup your go path
* Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Getting started
* Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
* Configure ssh-keys for github.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github profile, please follow steps in 
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further.

* Creating the executable
In order to create the ubiquity provisioner binary we need to start by getting the repository.
Clone the repository and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity-k8s.git
cd ubiquity-k8s
./scripts/build_provisioner
```
Newly built binary (provisioner) will be in bin directory.

### Running the Ubiquity dynamic provisioner
```bash
./bin/provisioner -config <configFile> -kubeconfig <kubeConfigDir> -provisioner <provisionerName> -retries=<number>
```
where:
* provisioner: Name of the dynamic provisioner (this will be used by the storage classes)
* configFile: Configuration file to use (defaults to `./ubiquity-client.conf`)
* kubeconfig: Local kubernetes configuration
* reties: Number of attempts for creating/deleting PVs for a given PVC

Example:
```bash
./bin/provisioner -provisioner=ubiquity/flex -config=./ubiquity-client.conf -kubeconfig=$HOME/.kube/config -retries=1
```
### Configuring the Ubiquity details

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity service will
look for a file named `ubiquity-client.conf` for its configuration.

The following snippet shows a sample configuration file:

```toml
logPath = "/tmp"  # The Ubiquity provisioner will write logs to file "ubiquity.log" in this path.
backend = "spectrum-scale" # Backend name

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening

```


### Usage example:
In order to test the dynamic provisioner create a storage class:
```bash
kubectl create -f deploy/storage_class_fileset.yml
```

The class is referring to `ubiquity/flex` as its provisioner. So this provisioner should be up and running in order to be able to dynamically create volumes.
`filesystem` parameter refers to the name of the filesystem to be used by the dynamic provisioner to create the volume. `backend` parameter is used to select the backend used by the system.
The `type` parameter is used to specify the type of volumes to be provisioned by spectrum-scale backend.

The following snippet shows a sample persistent volume claim for using dynamic provisioning:
```bash
kubectl create -f deploy/pvc_fileset.yml
```
The claim is referring to `spectrum-scale-fileset` as the storage class to be used.
A persistent volume should be dynamically created and bound to the claim.


# Ubiquity FlexVolume Cli for k8s

Ubiquity flexvolume cli provides access to persistent storage, via [ubiquity](https://github.com/IBM/ubiquity) service, within kubernetes

* Creating the executable
In order to create the ubiquity flexvolume binary we need to start by getting the repository.
Clone the repository (if you haven't done that yet) and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity-k8s.git
cd ubiquity-k8s
./scripts/build_flex_driver
```
 Newly built binary (ubiquity) will be in bin directory.

# Using the plugin
Install the ubiquity binary on all nodes in the kubelet plugin path along with its configuration file.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity/ubiquity
```

You can use the following commands to create and install the binary in the right location.

```bash
./scripts/build_flex_driver
./scripts/setup
```

In order to test the flexvolume within kubernetes, you can create the following pod based on the file in deploy folder (this suppose that you already used the dynamic provisioner to create the storageclass and the claim):
```bash
kubectl create -f deploy/pod.yml
```

#### Running Tests:
In order to test the dynamic provisioner, please refer to `scripts/run_acceptance.sh` file.
In order to test the flex driver, please refer to `scripts/flex_smoke_test.sh` file.
