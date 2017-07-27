# IBM Spectrum Scale

* [Deployment Prerequisities](#deployment-prerequisities)
* [Configuration](#configuring-ubiquity-docker-plugin-with-ubiquity-and-spectrum-scale)
* [Volume Creation](#volume-creation-using-spectrum-scale-storage-system)
  * [Fileset Volumes](#creating-fileset-volumes)
  * [Independent Fileset Volumes](#creating-independent-fileset-volumes)
  * [Lightweight Volumes](#creating-lightweight-volumes)
  * [Fileset with Quota Volumes](#creating-fileset-with-quota-volumes)

## Deployment Prerequisities
 * Spectrum-Scale - Ensure the Spectrum Scale client (NSD client) is installed and part of a Spectrum Scale cluster.
 * NFS - Ensure hosts support mounting NFS file systems.

## Ubiquity Dynamic Provisioner 

 
## Volume Creation usage
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

### Configuring Ubiquity Docker Plugin with Ubiquity and Spectrum Scale
 
 In case you need Spectrum Scale with NFS connectivity add this 2 lines into the ubiquity-client.conf file.
 
```toml
[SpectrumNfsRemoteConfig]
ClientConfig = "192.168.1.0/24(Access_Type=RW,Protocols=3:4,Transports=TCP:UDP)"
```
Where the ClientConfig contains the CIDR that the node where the volume will be mounted belongs to.
