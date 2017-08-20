# IBM Block Storage System via Spectrum Control Base Edition

IBM block storage can be used as persistent storage for Kubernetes via Ubiquity service.
Ubiquity communicates with the IBM storage systems through [IBM Spectrum Control Base Edition](https://www.ibm.com/support/knowledgecenter/en/STWMS9) (SCBE) 3.2.0. SCBE creates a storage profile (for example, gold, silver or bronze) and makes it available for Ubiquity FlexVolume and Ubiquity Dynamic Provisioner.
Available IBM block storage systems for Ubiquity FlexVolume and Ubiquity Dynamic Provisioner are listed in the [Ubiquity Service](https://github.com/IBM/ubiquity/).

This procedure explains how to configure [Ubiquity FlexVolume](#ubiquity-flexvolume-driver-cli) using SCBE. In addition, it provides [usage examples](#usage-example-for-ubiquity-dynamic-provisioner-and-flexvolume) for FlexVolume and Dynamic Provisioner.
Note : The Ubiquity Dynamic Provisioner configuration is described in the [README](README.md) file.


# Ubiquity FlexVolume Driver CLI

Perform the following installation and configuration procedures on each node in the Kubernetes cluster that requires access to Ubiquity volumes.


### 1. Installing connectivity packages
The FlexVolume supports FC or iSCSI connectivity to the storage systems.

  * RHEL, SLES

```bash
   sudo yum -y install sg3_utils
   sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
```

### 2. Configuring multipathing
The FlexVolume requires multipath devices. Configure the `multipath.conf` file according to the storage system requirements.
  * RHEL, SLES

```bash
   yum install device-mapper-multipath
   sudo modprobe dm-multipath

   cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
   systemctl start multipathd
   systemctl status multipathd  # Make sure its active
   multipath -ll  # Make sure no error appear.
```

### 3. Configure storage system connectivity
  *  Verify that the hostname of the Kubernetes node is defined on the relevant storage systems with the valid WWPNs or IQN of the node. The hostname on the storage system must be the same as the output of `hostname` command on the Kubernetes node. Otherwise, you will not be able to run stateful containers.

  *  For iSCSI, discover and log in to the iSCSI targets of the relevant storage systems:

```bash
   iscsiadm -m discoverydb -t st -p ${storage system iSCSI portal IP}:3260 --discover   # To discover targets
   iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login              # To log in to targets
```

### 4. Configuring Ubiquity FlexVolume for SCBE

The ubiquity-k8s-flex.conf must be created in the /etc/ubiquity directory. Configure the Ubiquity FlexVolume by editing the file, as illustrated below.

Just make sure backends set to "scbe".

 ```toml
 logPath = "/var/tmp"  # The Ubiquity FlexVolume will write logs to file "ubiquity-k8s-flex.log" in this path.
 backends = ["scbe"] # Backend name, such as scbe or spectrum-scale.
 logLevel = "info" # Optional parameter. Possible values are debug, info or error. Default is "info".


 [UbiquityServer]
 address = "IP"  # IP/hostname of the Ubiquity service
 port = 9999     # TCP port on which the Ubiquity service is listening
 ```
  * Verify that the logPath, exists on the host to enable the FlexVolume to run properly.

<br>
<br>
<br>
<br>


# Usage example for Ubiquity Dynamic Provisioner and FlexVolume

## Basic flow for running a stateful container with Ubiquity volume
Flow overview:
1. Create a StorageClass `gold` that refers to SCBE storage service `gold` with `xfs` as a file system type.
2. Create a PVC `pvc1` that uses the StorageClass `gold`.
3. Create a Pod `pod1` with container `container1` that uses PVC `pvc1`.
3. Start I/Os into `/data/myDATA` in `pod1\container1`.
4. Delete the `pod1` and then create a new `pod1` with the same PVC and verify that the file `/data/myDATA` still exists.
5. Delete the `pod1` `pvc1`, `pv` and storage class `gold`.

Relevant yml files (`storage_class_gold.yml`, `pvc1.yml` and `pod1.yml`):
```bash
#> cat storage_class_gold.yml pvc1.yml cat pod1.yml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: "gold"      # Storage Class name
  annotations:
   storageclass.beta.kubernetes.io/is-default-class: "true"  # Optional parameter. Set this Storage Class as the default
provisioner: "ubiquity/flex"  # Ubiquity provisioner name
parameters:
  profile: "gold"   # SCBE storage service name
  fstype: "xfs"     # Optional parameter. Possible values are ext4 or xfs. Default is configured on Ubiquity server
  backend: "scbe"   # Backend name for IBM block storage provisioning

kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: "pvc1"      # PVC name
  annotations:
    volume.beta.kubernetes.io/storage-class: "gold"  # The Storage Class name for the PVC
spec:
  accessModes:
    - ReadWriteOnce # Currently, Ubiquity scbe backend supports ReadWriteOnce mode only
  resources:
    requests:
      storage: 1Gi  # Size in Gi. Default size is configured on Ubiquity server

kind: Pod
apiVersion: v1
metadata:
  name: pod1          # Pod name
spec:
  containers:
  - name: container1  # Container name
    image: midoblgsm/kubenode
    volumeMounts:
      - name: vol1
        mountPath: "/data"  # mountpoint for vol1(pvc1)
  restartPolicy: "Never"
  volumes:
    - name: vol1
      persistentVolumeClaim:
        claimName: pvc1

```

Running the basic flow:
```bash
#> kubectl create -f storage_class_gold.yml -f pvc1.yml -f pod1.yml
storageclass "gold" created
persistentvolumeclaim "pvc1" created
pod "pod1" created

#### Wait for PV to be created and pod1 to be in the Running state...

#> kubectl get storageclass gold
NAME             TYPE
gold (default)   ubiquity/flex

#> kubectl get pvc pvc1
NAME      STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
pvc1      Bound     pvc-ba09bf4c-80ab-11e7-a42b-005056a46c49   1Gi        RWO           2m

#> kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM          REASON    AGE
pvc-ba09bf4c-80ab-11e7-a42b-005056a46c49   1Gi        RWO           Delete          Bound     default/pvc1             2m

#> kubectl get pod pod1
NAME      READY     STATUS    RESTARTS   AGE
pod1      1/1       Running   0          2m

#> kubectl exec pod1 -c container1  -- bash -c "df -h /data"
Filesystem          Size  Used Avail Use% Mounted on
/dev/mapper/mpathdo  951M   33M  919M   4% /data

#> kubectl exec pod1  -c container1 -- bash -c "dd if=/dev/zero of=/data/myDATA bs=10M count=1"

#> kubectl exec pod1  -c container1 -- bash -c "ls -l /data/myDATA"
-rw-r--r--. 1 root root 10485760 Aug 14 04:54 /data/myDATA

#> kubectl delete -f pod1.yml
pod "pod1" deleted

#### Wait for pod1 deletion...

#> kubectl get pod pod1
Error from server (NotFound): pods "pod1" not found

#> kubectl create -f pod1.yml
pod "pod1" created

#### Wait for pod1 to be in the Running state...

#> kubectl get pod pod1
NAME      READY     STATUS    RESTARTS   AGE
pod1      1/1       Running   0          2m

#### Verify the /data/myDATA still exist
#> kubectl exec pod1  -c container1 -- bash -c "ls -l /data/myDATA"
-rw-r--r--. 1 root root 10485760 Aug 14 04:54 /data/myDATA

### Delete pod1, pvc1, pv and the gold storage class
#> kubectl delete -f pod1.yml -f pvc1.yml -f storage_class_gold.yml
pod "pod1" deleted
persistentvolumeclaim "pvc1" deleted
storageclass "gold" deleted
```

<br>
<br>


## Basic flow breakdown
This section describes separate steps of the generic flow in greater detail.


### Creating a Storage Class
For example, to create a Storage Class named `gold` that refers to an SCBE storage service, such as a pool from IBM FlashSystem A9000R with QoS capability, and with the `xfs` file system type. As a result, every volume from this storage class will be provisioned on the `gold` SCBE service and will be initialized with `xfs` file system.
```bash
#> cat storage_class_gold.yml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: "gold"                 # Storage Class name
  annotations:
   storageclass.beta.kubernetes.io/is-default-class: "true" # Optional parameter. Set this the storage class as the default
provisioner: "ubiquity/flex"   # Ubiquity provisioner name
parameters:
  profile: "gold"              # SCBE storage service name
  fstype: "xfs"                # Optional parameter. Possible values are ext4 or xfs. Default is configured on the Ubiquity server
  backend: "scbe"              # Backend name for IBM block storage provisioning

#> kubectl create -f storage_class_gold.yml
storageclass "gold" created
```

List the newly created Storage Class:
```bash
#> kubectl get storageclass gold
NAME             TYPE
gold (default)   ubiquity/flex
```

### Creating a PersistentVolumeClaim
To create a PVC `pvc1` with size `1Gi` that uses the `gold` Storage Class:
```bash
#> cat pvc1.yml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: "pvc1"    # PVC name
  annotations:
    volume.beta.kubernetes.io/storage-class: "gold"  # The storage class name for the PVC
spec:
  accessModes:
    - ReadWriteOnce  # Currently, Ubiquity scbe backend supports ReadWriteOnce mode only
  resources:
    requests:
      storage: 1Gi  # Size in Gi. Default size is configured on Ubiquity server

#> kubectl create -f pvc1.yml
persistentvolumeclaim "pvc1 created
```

Ubiquity Dynamic Provisioner automatically creates a PersistentVolume (PV) and binds it to the PVC. The PV name will be PVC-ID. The volume name on the storage will be `u_[ubiquity-instance]_[PVC-ID]`. Note: [ubiquity-instance] is set in the Ubiquity server configuration file.

List a PersistentVolumeClaim and PersistentVolume
```bash
#> kubectl get pvc
NAME   STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
pvc1   Bound     pvc-254e4b5e-805d-11e7-a42b-005056a46c49   1Gi        RWO           1m

#> kubectl get pv
NAME                                       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS    CLAIM          REASON    AGE
pvc-254e4b5e-805d-11e7-a42b-005056a46c49   1Gi        RWO           Delete          Bound     default/pvc1             8s
```

Display the additional PV information, such as volume WWN, its location on the storage system etc:
```bash
#> kubectl get -o json pv pvc-254e4b5e-805d-11e7-a42b-005056a46c49 | grep -A15 flexVolume
        "flexVolume": {
            "driver": "ibm/ubiquity",
            "options": {
                "LogicalCapacity": "1000000000",
                "Name": "u_PROD_pvc-254e4b5e-805d-11e7-a42b-005056a46c49",
                "PhysicalCapacity": "1023410176",
                "PoolName": "gold-pool",
                "Profile": "gold",
                "StorageName": "A9000 system1",
                "StorageType": "2810XIV",
                "UsedCapacity": "0",
                "Wwn": "6001738CFC9035EB0000000000CCCCC5",
                "fstype": "xfs",
                "volumeName": "pvc-254e4b5e-805d-11e7-a42b-005056a46c49"
            }
        },
```

### Create a Pod with an Ubiquity volume
The creation of a Pod/Deployment causes the FlexVolume to:
* Attach the volume to the host
* Rescan and discover the multipath device of the new volume
* Create xfs or ext4 filesystem on the device (if filesystem does not exist on the volume)
* Mount the new multipath device on /ubiquity/[WWN of the volume]
* Create a symbolic link /var/lib/kubelet/pods/[POD-ID]/volumes/ibm~ubiquity-k8s-flex/[PVC-ID] -> /ubiquity/[WWN of the volume]

For example, to create a Pod `pod1` that uses the PVC `pvc1` that was already created:
```bash
#> cat pod1.yml
kind: Pod
apiVersion: v1
metadata:
  name: pod1          # Pod name
spec:
  containers:
  - name: container1  # Container name
    image: midoblgsm/kubenode
    volumeMounts:
      - name: vol1
        mountPath: "/data"  # Mountpoint for the vol1(pvc1)
  restartPolicy: "Never"
  volumes:
    - name: vol1
      persistentVolumeClaim:
        claimName: pvc1

#> kubectl create -f pod1.yml
pod "pod1" created
```

To display the newly created `pod1` and write data to the persistent volume of `pod1`:
```bash
#> kubectl get pod pod1
NAME      READY     STATUS    RESTARTS   AGE
pod1      1/1       Running   0          16m

#### Wait for pod1 to be in the Running state...

#> kubectl exec pod1 -c container1  -- bash -c "df -h /data"
Filesystem          Size  Used Avail Use% Mounted on
/dev/mapper/mpathi  951M   33M  919M   4% /data

#> kubectl exec pod1 -c container1  -- bash -c "mount | grep /data"
/dev/mapper/mpathi on /data type xfs (rw,relatime,seclabel,attr2,inode64,noquota)

#> kubectl exec pod1 touch /data/FILE
#> kubectl exec pod1 ls /data/FILE
File

#> kubectl describe pod pod1| grep "^Node:" # Where kubernetes deploy and run the Pod1
Node:		k8s-node1/[IP]
```

To display the newly attached volume on the minion node, log in to the minion that has the running pod and run the following commands:
```bash
#> multipath -ll
mpathi (36001738cfc9035eb0000000000cc2bc5) dm-12 IBM     ,2810XIV
size=954M features='1 queue_if_no_path' hwhandler='0' wp=rw
`-+- policy='service-time 0' prio=1 status=active
  |- 3:0:0:1 sdb 8:16 active ready running
  `- 4:0:0:1 sdc 8:32 active ready running

#> df | egrep "ubiquity|^Filesystem"
Filesystem                       1K-blocks    Used Available Use% Mounted on
/dev/mapper/mpathi                  973148   32928    940220   4% /ubiquity/6001738CFC9035EB0000000000CC2BC5

#> mount |grep ubiquity
/dev/mapper/mpathi on /ubiquity/6001738CFC9035EB0000000000CC2BC5 type xfs (rw,relatime,seclabel,attr2,inode64,noquota)

#> ls -l /var/lib/kubelet/pods/*/volumes/ibm~ubiquity-k8s-flex/*
lrwxrwxrwx. 1 root root 42 Aug 13 22:41 pvc-254e4b5e-805d-11e7-a42b-005056a46c49 -> /ubiquity/6001738CFC9035EB0000000000CC2BC5
```

### Deleting a Pod
The Kuberenetes delete Pod command:
* Removes symbolic link /var/lib/kubelet/pods/[POD-ID]/volumes/ibm~ubiquity-k8s-flex/[PVC-ID] -> /ubiquity/[WWN of the volume]
* Unmounts the new multipath device on /ubiquity/[WWN of the volume]
* Removes the multipath device of the volume
* Detaches (unmaps) the volume from the host
* Rescans with cleanup mode to remove the physical device files of the detached volume

For example:
```bash
#> kubectl delete pod pod1
pod "pod1" deleted
```

### Removing a volume
Removing the PVC deletes the PVC and its PV.

For example:
```bash
#> kubectl delete -f pvc1.yml
persistentvolumeclaim "pvc1" deleted
```

### Removing a Storage Class
For example:
```bash
#> kubectl delete -f storage_class_gold.yml
storageclass "gold" deleted
```

# Troubleshooting
### Server error
If the `bad status code 500 INTERNAL SERVER ERROR` error is displayed, check the `/var/log/sc/hsgsvr.log` log file on the SCBE node for explanation.

### Updating the volume on the storage side
Do not change a volume on a storage system itself, use `kubectl` command instead.
Any volume operation on the storage itself, requires a manual action on the minion (kublet node). For example, if you unmap a volume directly from the storage, you must clean up the multipath device of this volume and rescan the operating system on the minion (kubelet node).

### An attached volume cannot be attached to different host
A volume can be used only by one node at a time. In order to use a volume on different node, you must stop the Pod that uses the volume and then start a new Pod with the volume on different host.

### Cannot delete volume attached to a host
You cannot delete volume that is currently attached to a host. Any attempt will result in the `Volume [vol] already attached to [host]` error message.
If volume is not attached to any host, but this error is still displayed, run a new container, using this volume, then stop the container and remove the volume to delete it.
