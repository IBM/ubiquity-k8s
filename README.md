# Ubiquity-k8s
This projects contains the needed components to manage persistent storage for kubernetes through [Ubiquity](https://github.com/IBM/ubiquity) service.
The repository contains two components that could be used separately or combined according to the requirements:
- Ubiquity Dynamic Provisioner
- Ubiquity Flex Driver

This code is provided "AS IS" and without warranty of any kind.  Any issues will be handled on a best effort basis.

### General Prerequesites
* Functional [kubernetes]() environment (v1.5.x is required for flexvolume support, v1.6.x is not yet supported for the flexvolume)
* [Ubiquity](https://github.com/IBM/ubiquity) service must be running
* Install [golang](https://golang.org/) and setup your go path
* Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* The correct storage software must be installed and configured on each of the kubernetes nodes (minions). For example:
  * Spectrum-Scale - Ensure the Spectrum Scale client (NSD client) is installed and part of a Spectrum Scale cluster.
  * NFS - Ensure hosts support mounting NFS file systems.
* Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH

## Ubiquity Dynamic Provisioner 

Ubiquity Dynamic Provisioner facilitates creation and deletion of persistent storage in Kubernetes through use of the [Ubiquity](https://github.com/IBM/ubiquity) service.
  
### Download and build the code

```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
* Configure ssh-keys for github.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github profile, please follow steps in
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further.

* Creating the executable
In order to create the Ubiquity provisioner binary we need to start by getting the repository.
Clone the repository and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity-k8s.git
cd ubiquity-k8s
./scripts/build_provisioner
```
Newly built binary (provisioner) will be in bin directory.

### Configuration

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity service will
look for a file named `ubiquity-client.conf` for its configuration.

The following snippet shows a sample configuration file:

```toml
logPath = "/tmp/ubiquity"  # The Ubiquity provisioner will write logs to file "ubiquity-provisioner.log" in this path.
backend = "spectrum-scale" # Backend name

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening

```

If you need to use spectrum-scale-nfs backend, you need also to add the spectrum-scale nfs configuration:

```toml
[SpectrumNfsRemoteConfig]
ClientConfig = "192.168.1.0/24(Access_Type=RW,Protocols=3:4,Transports=TCP:UDP)"
```

Where the ClientConfig contains the CIDR that the node where the volume will be mounted belongs to.

### Two Options to Install and Run

#### Option 1: systemd
This option assumes that the system that you are using has support for systemd (e.g., ubuntu 14.04 does not have native support to systemd, ubuntu 16.04 does.)
Please note that the script will try to start the service as user `ubiquity`. The dynamic provisioner can run under any user, so if you prefer to use a different user please change the script to use the right user. Or create the user ubiquity as described in [Ubiquity documentation](https://github.com/IBM/ubiquity).

1) Change into the  ubiquity-k8s/scripts directory and run the following command:
```bash
./setup_provisioner
```

This will copy provisioner binary to /usr/bin, ubiquity-client-k8.conf and ubiquity-provisioner.env to  /etc/ubiquity location.  It will also enable ubiquity-provisioner service.

2) Make appropriate changes to /etc/ubiquity/ubiquity-client-k8.conf  e.g. server ip/port 

3) Edit /etc/ubiquity/ubiquity-provisioner.env to add/remove command line option to dynamic provisioner.

4) Start and stop the Ubiquity Kubernetes Dynamic Provisioner service the following command
```bash
systemctl start/stop/restart ubiquity-provisioner
```

#### Option 2: Manual
On any host in the cluster that can communicate with the Kubernetes admin node (although its probably simplest to run it on the same node as the Ubiquity server), run the following command

```bash
./bin/provisioner -config <configFile> -kubeconfig <kubeConfigDir> -provisioner <provisionerName> -retries=<number>
```
where:
* provisioner: Name of the dynamic provisioner (this will be used by the storage classes)
* configFile: Configuration file to use (defaults to `./ubiquity-client.conf`)
* kubeconfig: Local kubernetes configuration
* retries: Number of attempts for creating/deleting PVs for a given PVC

Example:
```bash
./bin/provisioner -provisioner=ubiquity/flex -config=./ubiquity-client.conf -kubeconfig=$HOME/.kube/config -retries=1
```

### Testing
In order to test the dynamic provisioner, please refer to `scripts/run_acceptance.sh` file.

### Available Storage Classes
These storage classes are described in the YAML files in `deploy` folder:
* spectrum-scale-fileset - described in `deploy/storage_class_fileset.yml`, it allows the dynamic provisioner to create volumes out of Spectrum Scale filesets.
* spectrum-scale-fileset-lightweight - described in `deploy/storage_class_lightweight.yml`, it allows the dynamic provisioner to create volumes out of sub-directories of filesets.

* spectrum-scale-fileset-nfs - described in `deploy/storage_class_fileset_nfs.yml`, it allows the dynamic provisioner to create volumes out of Spectrum Scale filesets based on NFS.

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


## Ubiquity FlexVolume CLI 

Ubiquity FlexVolume CLI supports attaching and detaching volumes on persistent storage using the [Ubiquity](https://github.com/IBM/ubiquity) service.

### Download and build the code

In order to create the Ubiquity flexvolume binary we need to start by getting the repository.
Clone the repository (if you haven't done that yet) and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity-k8s.git
cd ubiquity-k8s
./scripts/build_flex_driver
```
Newly built binary (ubiquity) will be in bin directory.

### Using the plugin
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

### Testing

In order to test the flex driver, please refer to `scripts/flex_smoke_test.sh` file.

In order to test the flexvolume within kubernetes, you can create the following pod based on the file in deploy folder (this suppose that you already used the dynamic provisioner to create the storageclass and the claim):
```bash
kubectl create -f deploy/pod.yml
```

## Suggestions and Questions

For any questions, suggestions, or issues, please use github.
