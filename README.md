# Ubiquity-k8s
This projects contains the needed components to manage persistent storage for kubernetes through [Ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service.
The repository contains mainly two components that could be used separately or combined according to the requirements:
- Ubiquity dynamic provisioner
- Ubiquity flex driver

# Ubiquity dynamic provisioner for k8s

Ubiquity provisioner facilitates creation and deletion of persistent storage, via [ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service, within kubernetes 

### Prerequesites
* Functional [kubernetes]() environment (v1.5.0 is required for flexvolume support)
* Spectrum-Scale client must be installed on the nodes
* [Ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service must be running
* Install [golang](https://golang.org/) and setup your go path
* Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Getting started
* Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH 
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
* Configure ssh-keys for github.ibm.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github.ibm profile, please follow steps in 
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further. 

* Creating the executable
In order to create the ubiquity provisioner binary we need to start by getting the repository.
Clone the repository and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-provisioner.git
cd ubiquity-provisioner
./scripts/build
```
Newly built binary (ubiquity) will be in bin directory. 

### Running the Ubiquity dynamic provisioner
```bash
./bin/ubiquity -config <configFile> -kubeconfig <kubeConfigDir> -provisioner <provisionerName>
```
where:
* configFile: Configuration file to use (defaults to `./ubiquity.conf`)

### Configuring the Ubiquity details

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity service will
look for a file named `ubiquity.conf` for its configuration.

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
kubectl create -f deploy/class.yml
```

The content of class.yml file is:
```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: "ubiquity-spectrum-scale"
  annotations:
   storageclass.beta.kubernetes.io/is-default-class: "true"
provisioner: "ubiquity/flex"
parameters:
  filesystem: "gold"
```

The class is refering to `ubiquity/flex` as its provisioner. So this provisioner should be up and running in order to be able to dynamically create volumes.
`filesystem` parameter refers to the name of the filesystem to be used by the dynamic provisioner to create the volume.

The following snippet shows a sample persistent volume claim for using dynamic provisioning:
```bash
kubectl create -f deploy/claim.yml
```
The claim is referring to `ubiquity-spectrum-scale` as the storage class to be used.
 ```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: "ubiquity-claim"
  annotations:
    volume.beta.kubernetes.io/storage-class: "ubiquity-spectrum-scale"
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
```

A persistent volume should be created and bound to the claim.


# Ubiquity FlexVolume Cli for k8s

Ubiquity flexvolume cli provides access to persistent storage, via [ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service, within kubernetes

* Creating the executable
In order to create the ubiquity flexvolume binary we need to start by getting the repository.
Clone the repository (if you haven't done that yet) and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-k8s.git
cd ubiquity-k8s
./scripts/build_provisioner
```
 Newly built binary (provisioner) will be in bin directory. 

# Using the plugin
Install the ubiquity binary on all nodes in the kubelet plugin path along with its configuration file.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/kubernetes.io~ubiquity/ubiquity
```

You can use the following commands to create and install the binary in the right location.

```bash
./scripts/build_flex_driver
./scripts/setup
```

# Driver invocation model
Init:
```bash
./ubiquity init
```
Attach:
```bash
./ubiquity attach <json options>
```
Detach:
```bash
./ubiquity detach <mount device>
```
Mount:
```bash
./ubiquity mount <target mount dir> <mount device> <json options>
```
Unmount:
```bash
./ubiquity unmount <mount dir>
```


# Driver output
```json
{
    "status": "<Success/Failure>",
    "message": "<Reason for success/failure>",
    "device": "<Path to the device attached. This field is valid only for attach calls>"
 }
 ```

# Usage example:
In order to test the flexvolume within kubernetes, you can create the following pod. You can use the pod description in test folder:
```bash
kubectl create -f test/ubiquity.yml
```

 ```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: ubiquity
    image: midoblgsm/kubenode
    volumeMounts:
    - name: ubiquity
      mountPath: /data
    ports:
    - containerPort: 8787
  volumes:
  - name: ubiquity
    flexVolume:
      driver: "kubernetes.io/ubiquity"
      options:
        volumeID: "vol1"
        size: "100m"
        filesystem: "gold"
 ```
