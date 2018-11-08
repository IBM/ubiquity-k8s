# Ubiquity
* [(PRODUCTNAME)](https://<PRODUCTURL>) is ... brief sentence regarding product
* Add "-Beta" as suffix if beta version - beta versions are generally < 1.0.0
* Don't include versions of charts or products

## Introduction
This chart bootstraps all Ubiquity components deployment on a Kubernetes cluster using the Helm package manager.

## Chart Details
This chart includes:
* A Ubiquity server Deployment used as the server of Kubernetes Dynamic Provisioner and FlexVolume
* A Ubiquity database Deployment used to store the persistent data of Ubiquity server
* A Kubernetes Dynamic Provisioner Deployment allows storage volumes to be created on-demand, using Kubernetes storage classes based on Spectrum Connect storage services.
* A Kubernetes FlexVolume DaemonSet enables the users to attach and mount storage volumes into a pod within a Kubernetes node.

## Prerequisites
* Kubernetes Level - indicate if specific APIs must be enabled (i.e. Kubernetes 1.6 with Beta APIs enabled)
* PersistentVolume requirements (if persistence.enabled) - PV provisioner support, StorageClass defined, etc. (i.e. PersistentVolume provisioner support in underlying infrastructure with ibmc-file-gold StorageClass defined if persistance.enabled=true)
* Simple bullet list of CPU, MEM, Storage requirements
* Even if the chart only exposes a few resource settings, this section needs to inclusive of all / total resources of all charts and subcharts.


## Resources Required
* Describes Minimum System Resources Required

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release --namespace ubiquity stable/ubiquity
```

The command deploys <Chart name> on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.


> **Tip**: List all releases using `helm list`

### Verifying the Chart
You can check the status by run:
```bash
$ helm status my-release
```

When all the status are fine, you can run sanity test by:
```bash
$ helm test my-release
```

### Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the <Ubiquity> chart and their default values.

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
