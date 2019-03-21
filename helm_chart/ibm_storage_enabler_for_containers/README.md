# IBM Storage Enabler for Containers

## Introduction
IBM Storage Enabler for Containers (ISEC) allows IBM storage systems to be used as persistent volumes for stateful applications running in Kubernetes clusters.
IBM Storage Enabler for Containers uses Kubernetes dynamic provisioning for creating and deleting volumes on IBM storage systems.
In addition, IBM Storage Enabler for Containers utilizes the full set of Kubernetes FlexVolume APIs for volume operations on a host.
The operations include initiation, attachment/detachment, mounting/unmounting etc..

## Chart Details
This chart includes:
* A Storage Enabler for Containers server for running Kubernetes Dynamic Provisioner and FlexVolume.
* A Storage Enabler for Containers database for storing the persistent data for the Enabler for Container service.
* A Kubernetes Dynamic Provisioner for creating storage volumes on-demand, using Kubernetes storage classes based on Spectrum Connect storage services or Spectrum Scale storage classes.
* A Kubernetes FlexVolume DaemonSet for attaching/detaching and mounting/unmounting storage volumes into a pod within a Kubernetes node.

## Prerequisites
Before installing the Helm chart for Storage Enabler for Containers in conjuction with IBM block storage:
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
5. If dedicated SSL certificates are required, see the Managing SSL certificates section in the IBM Storage Enabler for Containers user guide.
6. When using IBM Cloud Private with the Spectrum Virtualize Family products, use only hostnames for the Kubernetes cluster nodes, do not use IP addresses.

Prior to installing the Helm chart for Storage Enabler for Containers in conjunction with IBM Spectrum Scale:
1. Install and configure IBM Spectrum Scale, according to the application requirements.
2. Establish a proper communication link between Spectrum Scale Management GUI Address and Kubernetes cluster.
3. For each worker node:
   1. Install Spectum Scale client packages.
   2. Mount Spectrum Scale filesystem for persistent storage.
3. For each master node:
   1. Enable the attach/detach capability for the kubelet service.
   2. If the controller-manager is configured to run as a pod in your Kubernetes cluster, allow for event recording in controller-manager log file.

These configuration steps are mandatory and cannot be skipped. For detailed description, see the IBM Storage Enabler for Containers user guide on IBM Knowledge Center at https://www.ibm.com/support/knowledgecenter/SSCKLT.

## PodSecurityPolicy Requirements
This chart requires a PodSecurityPolicy to be bound to the target namespace prior to installation or to be bound to the current namespace during installation by setting "globalConfig.defaultPodSecurityPolicy.clusterRole". 

The predefined PodSecurityPolicy name: [`ibm-anyuid-hostpath-psp`](https://ibm.biz/cpkspec-psp) has been verified for this chart, if your target namespace is bound to this PodSecurityPolicy you can proceed to install the chart.
The predefined clusterRole name: ibm-anyuid-hostpath-clusterrole has been verified for this chart, if you use it you can proceed to install the chart.

You can also define a custom PodSecurityPolicy which can be used to finely control the permissions/capabilities needed to deploy this chart. You can enable this custom PodSecurityPolicy using the ICP user interface.

- From the user interface, you can copy and paste the following snippets to enable the custom PodSecurityPolicy
  - Custom PodSecurityPolicy definition:
    ```
    apiVersion: extensions/v1beta1
    kind: PodSecurityPolicy
    metadata:
      annotations:
        kubernetes.io/description: "This policy allows pods to run with 
          any UID and GID and any volume, including the host path.  
          WARNING:  This policy allows hostPath volumes.  
          Use with caution." 
      name: custom-psp
    spec:
      allowPrivilegeEscalation: true
      fsGroup:
        rule: RunAsAny
      requiredDropCapabilities: 
      - MKNOD
      allowedCapabilities:
      - SETPCAP
      - AUDIT_WRITE
      - CHOWN
      - NET_RAW
      - DAC_OVERRIDE
      - FOWNER
      - FSETID
      - KILL
      - SETUID
      - SETGID
      - NET_BIND_SERVICE
      - SYS_CHROOT
      - SETFCAP 
      runAsUser:
        rule: RunAsAny
      seLinux:
        rule: RunAsAny
      supplementalGroups:
        rule: RunAsAny
      volumes:
      - '*'
    ```
  - Custom ClusterRole for the custom PodSecurityPolicy:
    ```
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: custom-clusterrole
    rules:
    - apiGroups:
      - extensions
      resourceNames:
      - custom-psp
      resources:
      - podsecuritypolicies
      verbs:
      - use
    ```

## Resources Required
IBM Storage Enabler for Containers can be deployed on different operating systems, while provisioning storage from a variety of IBM arrays, as detailed in the release notes of the package.  

## Installing the Chart
To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release --namespace ubiquity stable/ibm-storage-enabler-for-containers
```

The command deploys <chart name> on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.


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
Verify that there are no persistent volumes (PVs) that have been created, using IBM Storage Enabler for Containers. 
To uninstall/delete the `my-release` release:

```bash
$ helm delete `my-release` --purge
```

The command removes the IBM Storage Enabler for Containers components associated with the Helm chart, metadata, user credentials, and other elements.

When the Helm chart is deleted, the first elements to be removed are the Enabler for Container database deployment and its PVC. If the `helm delete` command fails after several attempts, delete these entities manually before continuing. Then, verify that the Enabler for Container database deployment and its PVC are deleted, and complete the uninstall procedure by running
```
$ helm delete `my-release` --purge --no-hooks
```
## Configuration

The following table lists the configurable parameters of the <Ubiquity> chart and their default values.

| Parameter                  | Description                                     | Default                                                    |
| -----------------------    | ---------------------------------------------   | ---------------------------------------------------------- |
| `images.ubiquity`                                                                               | Image for ISEC server | `ibmcom/ibm-storage-enabler-for-containers:2.1.0` |
| `images.ubiquitydb`                                                                             | Image for ISEC database | `ibmcom/ibm-storage-enabler-for-containers-db:2.1.0` |
| `images.provisioner`                                                                            | Image for Kubernetes Dynamic Provisioner | `ibmcom/ibm-storage-dynamic-provisioner-for-kubernetes:2.1.0` |
| `images.flex`                                                                                   | Image for Kubernetes FlexVolume | `ibmcom/ibm-storage-flex-volume-for-kubernetes:2.1.0` |
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
| `ubiquityK8sFlex.flexLogDir`                                                       |Directory for storing the ubiquity-k8s-flex.log file.  | /var/log |
| `ubiquityK8sFlex.ubiquityIPaddress`                                                       |IP address of the ubiquity service object.  | |
| `globalConfig.logLevel`                                                       |Log level. Allowed values: debug, info, error.  | info |
| `globalConfig.sslMode`                                                       |SSL verification mode. Allowed values: require (No validation is required, the IBM Storage Enabler for Containers server generates self-signed certificates on the fly.) or verify-full (Certificates are provided by the user.).  | require |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart.

## Storage
IBM Storage Enabler for Containers allows IBM storage system volumes to be used for stateful applications running in Kubernetes clusters. The full list of supported IBM systems is detailed in the Enabler for Containers realease notes (see the Documentation section below).

## Troubleshooting
You can use the IBM Storage Enabler for Containers logs for problem identification. To collect and display logs, related to the different components of IBM Storage Enabler for Containers, use this script: 

```bash
./ubiquity_cli.sh -a collect_logs
```
The logs are kept in the `./ubiquity_collect_logs_MM-DD-YYYY-h:m:s` folder. The folder is placed in the directory, from which the log collection command was run.

## Limitations
* Only one type of IBM storage backend (block or file) can be configured on the same Kubernetes or ICP cluster.
* Only one instance of IBM Storage Enabler for Containers can be deployed in a Kubernetes cluster, serving all Kubernetes namespaces.
* None of the deployments under this chart  support scaling. Thus, their replica must be 1.

## Documentation
Full documentation set for IBM Storage Enabler for Containers is available on IBM Knowledge Center at https://www.ibm.com/support/knowledgecenter/SSCKLT.
