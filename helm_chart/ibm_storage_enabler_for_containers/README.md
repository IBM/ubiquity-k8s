# Ubiquity
* [(PRODUCTNAME)](https://<PRODUCTURL>) is ... brief sentence regarding product
* Add "-Beta" as suffix if beta version - beta versions are generally < 1.0.0
* Don't include versions of charts or products

## Introduction
IBM Storage Enabler for Containers allows IBM storage systems to be used as persistent volumes for stateful applications running in Kubernetes clusters.
Thus, the containers can be used with stateful microservices, such as database applications (MongoDB, PostgreSQL etc).
IBM Storage Enabler for Containers uses Kubernetes dynamic provisioning for creating and deleting volumes on IBM storage systems.
In addition, IBM Storage Enabler for Containers utilizes the full set of Kubernetes FlexVolume APIs for volume operations on a host.
The operations include initiation, attachment/detachment, mounting/unmounting etc..

## Chart Details
This chart includes:
* A Ubiquity server Deployment used as the server of Kubernetes Dynamic Provisioner and FlexVolume.
* A Ubiquity database Deployment used to store the persistent data of Ubiquity server.
* A Kubernetes Dynamic Provisioner Deployment for creation  storage volumes on-demand, using Kubernetes storage classes based on Spectrum Connect storage services.
* A Kubernetes FlexVolume DaemonSet for attaching and mounting storage volumes into a pod within a Kubernetes node.

## Prerequisites
Before installing the helm chart, verify following:
1. Install and configure IBM Spectrum Connect, according to the application requirements.
2. Establish a proper communication link between Spectrum Connect and Kubernetes cluster.
3. For each worker node:
   1. Install relevant Linux packages to ensure Fibre Channel and iSCSI connectivity.
   2. Configure Linux multipath devices on the host.
   3. Configure storage system connectivity.
   4. Make sure that the node kubelet service has the attach/detach capability enabled.
4. For each master node:
   1. Enable the attach/detach capability for the kubelet service.
   2. If the controller-manager is configured to run as a pod in your Kubernetes cluster, allow for event recording in controller-manager log file.
5. If dedicated SSL certificates are required, see the Managing SSL certificates section in the IBM Storage Enabler for Containers.
6. When using IBM Cloud Private with the Spectrum Virtualize Family products, use only hostnames for the Kubernetes cluster nodes, do not use IP addresses.

These configuration steps are mandatory and cannot be skipped. See the IBM Storage Enabler for Containers user guide for their detailed description.


## Resources Required
* Describes Minimum System Resources Required

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release --namespace ubiquity stable/ibm_storage_enabler_for_containers
```

The command deploys <Chart name> on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.


> **Tip**: List all releases using `helm list`

### Verifying the Chart
You can check the status by running:
```bash
$ helm status my-release
```

If all statuses are free of errors, you can run sanity test by:
```bash
$ helm test my-release
```

### Uninstalling the Chart

To uninstall/delete the `my-release` release:

```bash
$ helm delete my-release --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the <Ubiquity> chart and their default values.

| Parameter                  | Description                                     | Default                                                    |
| -----------------------    | ---------------------------------------------   | ---------------------------------------------------------- |
| `images.ubiquity`                                                                               | Image for Ubiquity server | `ibmcom/ibm-storage-enabler-for-containers:2.0.0` |
| `images.ubiquitydb`                                                                             | Image for Ubiquity database | `ibmcom/ibm-storage-enabler-for-containers-db:2.0.0` |
| `images.provisioner`                                                                            | Image for Kubernetes Dynamic Provisioner | `ibmcom/ibm-storage-dynamic-provisioner-for-kubernetes:2.0.0` |
| `images.flex`                                                                                   | Image for Kubernetes FlexVolume | `ibmcom/ibm-storage-flex-volume-for-kubernetes:2.0.0` |
| `spectrumConnect.connectionInfo.fqdn`                                                           | IP\FQDN of Spectrum Connect server. | ` ` |
| `spectrumConnect.connectionInfo.port`                                                           | Port of Spectrum Connect server. | ` ` |
| `spectrumConnect.connectionInfo.username`                                                       | Username defined for IBM Storage Enabler for Containers interface in Spectrum Connect. | ` ` |
| `spectrumConnect.connectionInfo.password`                                                       | Password defined for IBM Storage Enabler for Containers interface in Spectrum Connect. | ` ` |
| `spectrumConnect.connectionInfo.sslMode`                                                        | SSL verification mode. Allowed values: require (no validation is required) and verify-full (user-provided certificates) | `require` |
| `spectrumConnect.backendConfig.instanceName`                                                    | A prefix for any new volume created on the storage system | ` ` |
| `spectrumConnect.backendConfig.skipRescanIscsi`                                                 | Allowed values: true or false. Set to true if the nodes have FC connectivity | `false` |
| `spectrumConnect.backendConfig.DefaultStorageService`                                           | Default Spectrum Connect storage service to be used, if not specified by the storage class | ` ` |
| `spectrumConnect.backendConfig.newVolumeDefaults.fsType`                                        | File system type. Allowed values: ext4 or xfs | `ext4` |
| `spectrumConnect.backendConfig.newVolumeDefaults.size`                                          | The default volume size (in GB) if not specified by the user when creating a new volume | `1` |
| `spectrumConnect.backendConfig.dbPvConfig.ubiquityDbPvName`                                     | Ubiquity database PV name. For Spectrum Virtualize and Spectrum Accelerate, use default value "ibm-ubiquity-db". For DS8000 Family, use "ibmdb" instead and make sure UBIQUITY_INSTANCE_NAME_VALUE value length does not exceed 8 chars | `ibm-ubiquity-db`                                                        |
| `spectrumConnect.backendConfig.dbPvConfig.storageClassForDbPv.storageClassName`                 | Parameters to create the first Storage Class that also be used by ubiquity for ibm-ubiquity-db PVC | ` `|
| `spectrumConnect.backendConfig.dbPvConfig.storageClassForDbPv.params.spectrumConnectServiceName`| Storage Class profile parameter should point to the Spectrum Connect storage service name | ` ` |
| `spectrumConnect.backendConfig.dbPvConfig.storageClassForDbPv.params.fsType`                    | Storage Class file-system type, Allowed values: ext4 or xfs | `ext4` |
| `genericConfig.ubiquityIpAddress`                                                               | The IP address of the ubiquity service object | ` ` |
| `genericConfig.logging.logLevel`                                                                | Log level. Allowed values: debug, info, error | `info` |
| `genericConfig.logging.flexLogDir`                                                              | Flex log directory. If you change the default, then make the new path exist on all the nodes and update the Flex daemonset hostpath according | `/var/log` |
| `genericConfig.ubiquityDbCredentials.username`                                                  | Username for the deployment of ubiquity-db database. Note : Do not use the "postgres" username, because it already exists | ` ` |
| `genericConfig.ubiquityDbCredentials.password`                                                  | Password for the deployment of ubiquity-db database | ` ` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart.

## Storage
* Define how storage works with the workload
* Dynamic vs PV pre-created
* Considerations if using hostpath, local volume, empty dir
* Loss of data considerations
* Any special quality of service or security needs for storage

## Limitations
* Deployment limits - can you deploy more than once, can you deploy into different namespace
* List specific limitations such as platforms, security, replica's, scaling, upgrades etc.. - noteworthy limits identified
* List deployment limitations such as : restrictions on deploying more than once or into custom namespaces.
* Not intended to provide chart nuances, but more a state of what is supported and not - key items in simple bullet form.
* Does it support IBM Cloud Kubernetes Service in addition to IBM Cloud Private?

## Documentation
* Can have as many supporting links as necessary for this specific workload however don't overload the consumer with unnecessary information.
* Can be links to special procedures in the knowledge center.
