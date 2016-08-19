# Spectrum Scale FlexVolume Cli for kubernetes

Spectrum Scale flexvolume cli provides access to persistent storage, utilizing Spectrum Scale, within kubernetes

# Prerequesites

* GPFS client must be installed in the nodes
* Install [golang](https://golang.org/)

# Creating the executable
In order to create the spectrum binary we need to start by getting all the dependencies.
We suppose that [godep](https://github.com/tools/godep) is installed.

```bash
git clone git@github.ibm.com:almaden-containers/spectrum-flexvolume-cli.git
cd spectrum-flexvolume-cli
godep restore
go build -o output/spectrum
```

# Testing the plugin
Install the spectrum binary on all nodes in the kubelet plugin path.

Path for installing the plugin is:
```bash
/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~spectrum/spectrum
```
# Driver invocation model
Init:
```bash
spectrum init
```
Attach:
```bash
spectrum attach <json options>
```
Detach:
```bash
spectrum detach <mount device>
```
Mount:
```bash
spectrum mount <target mount dir> <mount device> <json options>
```
Unmount:
```bash
spectrum unmount <mount dir>
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
    - name: spectrum
      mountPath: /data
    ports:
    - containerPort: 80
  volumes:
  - name: spectrum
    flexVolume:
      driver: "ibm/spectrum"
      fsType: "ext4"
      options:
        volumeID: "gpfs1"
        size: "1000m"
        fileset: "filesetid"
 ```
