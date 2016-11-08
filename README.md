# Ubiquity FlexVolume Cli for kubernetes

Ubiquity flexvolume cli provides access to persistent storage, via [ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service, within kubernetes

### Prerequesites
* Functional [kubernetes]() environment (v1.4.0 is required for flexvolume support)
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
In order to create the ubiquity flexvolume binary we need to start by getting the repository.
Clone the repository and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-flexvolume.git
cd ubiquity-flexvolume
./scripts/build
```
 Newly built binary (ubiquity) will be in bin directory. 

# Using the plugin
Install the ubiquity binary on all nodes in the kubelet plugin path along with its configuration file.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/kubernetes.io~ubiquity/ubiquity
```

You can use the following command to create and install the binary in the right location.

```bash
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
