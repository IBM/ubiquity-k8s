# Ubiquity
* [(PRODUCTNAME)](https://<PRODUCTURL>) is ... brief sentence regarding product
* Add "-Beta" as suffix if beta version - beta versions are generally < 1.0.0
* Don't include versions of charts or products

## Introduction
IBM Storage Enabler for Containers (ISEC) allows IBM storage systems to be used as persistent volumes for stateful applications running in Kubernetes clusters.
Thus, the containers can be used with stateful microservices, such as database applications (MongoDB, PostgreSQL etc).
IBM Storage Enabler for Containers uses Kubernetes dynamic provisioning for creating and deleting volumes on IBM storage systems.
In addition, IBM Storage Enabler for Containers utilizes the full set of Kubernetes FlexVolume APIs for volume operations on a host.
The operations include initiation, attachment/detachment, mounting/unmounting etc..

## Chart Details
This chart includes:
* A Storage Enabler for Containers server deployment is used as the server for running Kubernetes Dynamic Provisioner and FlexVolume.
* A Storage Enabler for Containers database deployment is used to store the persistent data for Enabler for Container server.
* A Kubernetes Dynamic Provisioner deployment is for creation storage volumes on-demand, using Kubernetes storage classes based on Spectrum Connect storage services.
* A Kubernetes FlexVolume DaemonSet is used for attaching and mounting storage volumes into a pod within a Kubernetes node.

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

## PodSecurityPolicy Requirements
This chart requires a PodSecurityPolicy to be bound to the target namespace prior to installation or to be bound to the current namespace during installation by setting "globalConfig.defaultPodSecurityPolicy.clusterRole". 

The predefined PodSecurityPolicy name: ibm-anyuid-hostpath has been verified for this chart, if your target namespace is bound to this PodSecurityPolicy you can proceed to install the chart.
The predefined clusterRole name: ibm-anyuid-hostpath-clusterrole has been verified for this chart, if you use it you can proceed to install the chart.

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

This command fails if any of the following resources do not exist: serviceAccount, role, clusterRole, roleBinding, clusterRoleBinding with same name `ubiquity-helm-hook`. To continue the deletion, first create these entities.
> **Tip**: You can generate the manifests of these resources by running the "helm install --debug –dry-run" command with same release name, namespace and other values.

## Configuration

The following table lists the configurable parameters of the <Ubiquity> chart and their default values.

| Parameter                  | Description                                     | Default                                                    |
| -----------------------    | ---------------------------------------------   | ---------------------------------------------------------- |
| `images.ubiquity`                                                                               | Image for ISEC server | `ibmcom/ibm-storage-enabler-for-containers:2.0.0` |
| `images.ubiquitydb`                                                                             | Image for ISEC database | `ibmcom/ibm-storage-enabler-for-containers-db:2.0.0` |
| `images.provisioner`                                                                            | Image for Kubernetes Dynamic Provisioner | `ibmcom/ibm-storage-dynamic-provisioner-for-kubernetes:2.0.0` |
| `images.flex`                                                                                   | Image for Kubernetes FlexVolume | `ibmcom/ibm-storage-flex-volume-for-kubernetes:2.0.0` |
| `ubiquity.spectrumConnect.connectionInfo.fqdn`                                                           | IP address or FQDN of the Spectrum Connect server. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.port`                                                           |Communication port of the Spectrum Connect server. Default value is 8440. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.username`                                                       | Username defined for IBM Storage Enabler for Containers interface in Spectrum Connect. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.password`                                                       | Password defined for IBM Storage Enabler for Containers interface in Spectrum Connect. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.backendConfig.instanceName`                                                       |A prefix for any new volume created on the storage system. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.backendConfig.defaultStorageService`                                                       |Default Spectrum Connect storage service to be used, if not specified by the storage class. | ` ` |
| `ubiquity.spectrumConnect.connectionInfo.backendConfig.fsType`                                                       |File system type of a new volume, if not specified by the user in the storage class. Allowed values: ext4 or xfs. | ext4 |
| `ubiquity.spectrumConnect.connectionInfo.backendConfig.size`                                                       |Default volume size (in GB), if not specified by the user when creating a new volume. | 1 |
| `ubiquityDb.dbCredentials.username`                                                       |Username for the deployment of ubiquity-db database. Do not use the postgres username, because it already exists. |  |
| `ubiquityDb.dbCredentials.password`                                                       |Password for the deployment of ubiquity-db database. |  |
| `ubiquityDb.persistence.pvName`                                                       |Name of the persistent volume to be used for the ubiquity-db database. For the Spectrum Virtualize and Spectrum Accelerate storage systems, use the default value (ibm-ubiquity-db). For the DS8000 storage system, use a shorter value, such as (ibmdb). This is necessary because the DS8000 volume name length cannot exceed 8 characters. |  |
| `ubiquityDb.persistence.pvSize`                                                       |Default size (in GB) of the persistent volume to be used for the ubiquity-db database. | 20 |
| `ubiquityDb.persistence.useExistingPv`                                                       |Enabling the usage of an existing PV as the ubiquity-db database PV. Allowed values: True or False. | True |
| `ubiquityDb.persistence.storageClass.storageClassName`                                                       |Storage class name. The storage class parameters are used for creating an initial storage class for the ubiquity-db PVC. You can use this storage class for other applications as well. It is recommended to set the storage class name to be the same as the Spectrum Connect storage service name. | |
| `ubiquityDb.persistence.storageClass.existingStorageClass`                                                       |Enabling the usage of an existing storage class object if it exists. | |
| `ubiquityDb.persistence.storageClass.spectrumConnect.spectrumConnectServiceName`                                                       |Storage class profile, directing to the Spectrum Connect storage service name. | |
| `ubiquityDb.persistence.storageClass.spectrumConnect.fsType`                                                       |File system type for the storage class profile. Allowed values: ext4 or xfs.  | ext4 |


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
| `genericConfig.ubiquityDbCredentials.username`                                                  | Username for the deployment of ubiquity-db database. Do not use the "postgres" username, because it already exists | ` ` |
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
* Only one instance of IBM Storage Enabler for Containers can be deployed in a Kubernetes cluster.
*  None of the deployments under this chart  support scaling. Thus, their replica must be 1.

## Documentation
* Can have as many supporting links as necessary for this specific workload however don't overload the consumer with unnecessary information.
* Can be links to special procedures in the knowledge center.
