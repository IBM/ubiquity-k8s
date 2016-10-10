# Ubiquity FlexVolume Cli for kubernetes

Ubiquity flexvolume cli provides access to persistent storage, utilizing Spectrum Scale, within kubernetes

# Prerequesites

* GPFS client must be installed in the nodes
* Install [golang](https://golang.org/)

# Creating the executable
In order to create the spectrum binary we need to start by getting all the dependencies.
We suppose that [godep](https://github.com/tools/godep) is installed.

```bash
git clone git@github.ibm.com:almaden-containers/ubiquity-flexvolume.git
cd ubiquity-flexvolume
godep restore
go build -o output/ubiquity
```

# Testing the plugin
Install the ubiquity binary on all nodes in the kubelet plugin path.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity/ubiquity
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

# Usage example:

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
