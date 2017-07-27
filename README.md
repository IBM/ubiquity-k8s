# Ubiquity-k8s
This projects contains the needed components to manage persistent storage for kubernetes through [Ubiquity](https://github.com/IBM/ubiquity) service.
The repository contains two components that could be used separately or combined according to the requirements:
- Ubiquity Dynamic Provisioner
- Ubiquity Flex Driver

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Ubiquity Dynamic Provisioner 

Ubiquity Dynamic Provisioner facilitates creation and deletion of persistent storage in Kubernetes through use of the [Ubiquity](https://github.com/IBM/ubiquity) service.
  
### Installing the Ubiquity Dynamic Provisioner
Install and configure the Provisioner on one node in the Kubernetes cluster(minion or master)

### 1. Prerequisites
  * The Provisioner is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

  * The following sudoers configuration `/etc/sudoers` is required to run the Provisioner as root user: 
  
     ```
        Defaults !requiretty
     ```
     For non-root users, such as USER, configure the sudoers as follows: 

     ```
         USER ALL= NOPASSWD: /usr/bin/, /bin/, /usr/sbin/ 
         Defaults:%USER !requiretty
         Defaults:%USER secure_path = /sbin:/bin:/usr/sbin:/usr/bin
     ```

  TBD * The Provisioner  node must have access to the storage backends. Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.
   

### 2. Downloading and installing the plugin

* Download and unpack the application package.
```bash
mkdir -p /etc/ubiquity
cd /etc/ubiquity
curl -L https://github.com/IBM/ubiquity-k8s/releases/download/v0.4.0/ubiquity-k8s-provisioner-0.4.0.tar.gz | tar xf -
cp provisioner /usr/bin 
chmod u+x /usr/bin/ubiquity-k8s-provisioner
#chown USER:GROUP /usr/bin/ubiquity-k8s-provisioner   ### Run this command only a non-root user.
cp ubiquity-k8s-provisioner.service /usr/lib/systemd/system/ 
```
   * To run the plugin as non-root user, add the `User=USER` line under the [Service] item in the  `/usr/lib/systemd/system/ubiquity-k8s-provisioner.service` file.
   
   * Enable the plugin service.
   
```bash 
systemctl enable ubiquity-k8s-provisioner.service      
```

### 3. Configuring the plugin
Before running the plugin service, you must create and configure the `/etc/ubiquity/ubiquity-k8s-provisioner.conf` file, according to your storage system type.

Here is example of a configuration file that need to be set:
```toml
logPath = "/tmp/ubiquity"  # The Ubiquity provisioner will write logs to file "ubiquity-provisioner.log" in this path.
backend = "scbe" # Backend name such as scbe or spectrum-scale

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening

```

Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section in order to configure additional parameters that related.

TBD this section need to be move to dedicate SSc file
==============================
If you need to use spectrum-scale-nfs backend, you need also to add the spectrum-scale nfs configuration:

```toml
[SpectrumNfsRemoteConfig]
ClientConfig = "192.168.1.0/24(Access_Type=RW,Protocols=3:4,Transports=TCP:UDP)"
```
Where the ClientConfig contains the CIDR that the node where the volume will be mounted belongs to.
==============================

### 4. Running the plugin service
  * Run the service.
```bash
systemctl start ubiquity-docker-plugin    
```
  * Restart the Docker engine daemon on the host to let it discover the new plugin. 
```bash
service docker restart
```

TBD section below
==================
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
===========================================================

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
