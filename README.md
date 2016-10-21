# Ubiquity FlexVolume Cli for kubernetes

Ubiquity flexvolume cli provides access to persistent storage, via [ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service, within kubernetes

# Prerequesites


* Functional [kubernetes]() environment (v1.4.0 is required for flexvolume support)
* GPFS client must be installed in the nodes
* Ubiquity service must be running
* Install [golang](https://golang.org/) and setup your go path
* Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

# Creating the executable
In order to create the ubiquity flexvolume binary we need to start by getting the repository.
Clone the repository and build the binary using these commands.

```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-flexvolume.git
cd ubiquity-flexvolume.git
./bin/build
```

The build command will create a new folder `out`. It will also build the binary in this folder.

# Using the plugin
Install the ubiquity binary on all nodes in the kubelet plugin path.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity/ubiquity
```

You can use the following command to create and install the binary in the right location.

```bash
./bin/install
```

# Driver invocation model
Init:
```bash
ubiquity init
```
Attach:
```bash
ubiquity attach <json options>
```
Detach:
```bash
ubiquity detach <mount device>
```
Mount:
```bash
ubiquity mount <target mount dir> <mount device> <json options>
```
Unmount:
```bash
ubiquity unmount <mount dir>
```


# Driver output
```json
{
    "status": "<Success/Failure>",
    "message": "<Reason for success/failure>",
    "device": "<Path to the device attached. This field is valid only for attach calls>"
 }
 ```

You can also use the script in test folder to test the different actions.
The script will create a spectrum volume, attach it, mount it, unmount it, detach it and delete the volume.

```bash
./test/test.sh
```

# Usage example:
In order to test the flexvolume within kubernetes, you can create the following pod. You can use the pod description in test folder:
```bash
kubectl create -f test/spectrum.yml
```

 ```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - name: ubiquity
      mountPath: /data
    ports:
    - containerPort: 80
  volumes:
  - name: ubiquity
    flexVolume:
      driver: "ibm/ubiquity"
      options:
        volumeID: "gpfs1"
        size: "1000m"
        fileset: "filesetid"
 ```
