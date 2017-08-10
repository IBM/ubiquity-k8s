# IBM Block Storage System via Spectrum Control Base Edition

IBM block storage can be used as persistent storage for Kubernetes via Ubiquity service.
Ubiquity communicates with the IBM storage systems through [IBM Spectrum Control Base Edition](https://www.ibm.com/support/knowledgecenter/en/STWMS9) (SCBE) 3.2.0. SCBE creates a storage profile (for example, gold, silver or bronze) and makes it available for Ubiquity Docker volume plugin.
Avilable IBM block storage systems for Docker volume plugin are listed in the [Ubiquity Service](https://github.com/IBM/ubiquity/).

# Ubiquity Dynamic Provisioner 
There is no tuning needed for the Provisioner related to IBM Block Storage.

# Ubiquity FlexVolume CLI

## Configuring Kubernetes node(minion) for IBM block storage systems
Perform the following installation and configuration procedures on each node in the Kubernetes cluster that requires access to Ubiquity volumes.


#### 1. Installing connectivity packages 
The plugin supports FC or iSCSI connectivity to the storage systems.

  * RHEL, SLES
  
```bash
   sudo yum -y install sg3_utils
   sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
```

#### 2. Configuring multipathing 
The plugin requires multipath devices. Configure the `multipath.conf` file according to the storage system requirments.
  * RHEL, SLES
  
```bash
   yum install device-mapper-multipath
   sudo modprobe dm-multipath

   cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
   systemctl start multipathd
   systemctl status multipathd  # Make sure its active
   multipath -ll  # Make sure no error appear.
```

#### 3. Configure storage system connectivity
  *  Verify that the hostname of the Kubernetes node is defined on the relevant storage systems with the valid WWPNs or IQN of the node. The hostname on the storage system must be the same as the output of `hostname` command on the Docker node. Otherwise, you will not be able to run stateful containers.

  *  For iSCSI, discover and log in to the iSCSI targets of the relevant storage systems:

```bash
   iscsiadm -m discoverydb -t st -p ${storage system iSCSI portal IP}:3260 --discover   # To discover targets
   iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login              # To log in to targets
```

#### 5. Configuring Ubiquity FlexVolume for SCBE

The ubiquity-k8s-flex.conf must be created in the /etc/ubiquity directory. Configure the plugin by editing the file, as illustrated below.

Just make sure backends set to "scbe".
 
 ```toml
 logPath = "/var/tmp/ubiquity"                 # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.
 backend = "scbe" # Backend name such as scbe or spectrum-scale
 
 [UbiquityServer]
 address = "IP"  # IP/hostname of the Ubiquity Service
 port = 9999     # TCP port on which the Ubiquity Service is listening
 ```
  * Verify that the logPath, exists on the host before you start the plugin.
 
  

## Plugin usage example

### Basic flow for running a stateful container with Ubiquity volume
The basic flow is as follows:
1. Create a StorageClass that will refer to SCBE service.
2. Create a PVC that uses the StorageClass.
3. Create a Pod/Deployment to use the PVC.
3. Start I/Os into `/data/myDATA` inside the Pod.
4. Exit the Pod and then start a new Pod with the same PVC and validate that the file `/data/myDATA` still exists.
5. Clean up by exiting the Pod, removing the containers and deleting the PVC.


### Creating a StorageClass
Kubernetes Storage Class creation template:
```bash
#> kubectl create -f <path/to/storageClass.yml>
```
The storage class should refer to `ubiquity/flex` as its provisioner.
For example, to create a StorageClass named "gold" that refers to the SCBE storage service, such as a pool from IBM FlashSystem A9000R with QoS capability:

```bash
#> kubectl create -f deploy/scbe_volume_storage_class.yml

storageclass "gold" created
```

### Displaying the StorageClass
You can list the newly created StorageClass, using the following command:
```bash
#> kubectl get storageclass

NAME                               TYPE
gold (default)                     ubiquity/flex
```

### Creating a PersistentVolumeClaim
Kubernetes PVC creation template:
```bash
#> kubectl create -f <path/to/pvc.yml>
```

For example, to create a PVC that refers to the Gold StorageClass that we created:
```bash
#> kubectl create -f deploy/scbe_volume_pvc.yml

persistentvolumeclaim "scbe-accept-vol1" created
```

The PVC will be created and the dynamic provisioner will create a PersistentVolume (PV) and bind it to the PVC.
To list the PVC and PV that got created:

```bash
#> kubectl get pvc,pv
NAME                             STATUS    VOLUME                                      CAPACITY   ACCESSMODES   STORAGECLASS             AGE
pvc/scbe-accept-vol1              Bound    pvc-e6ea27f1-7d51-11e7-af04-0894ef20e599      20Gi         RWO          gold                  1m

NAME                                          CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS     CLAIM                                STORAGECLASS             REASON    AGE
pv/pvc-e6ea27f1-7d51-11e7-af04-0894ef20e599   1Gi         RWO           Delete          Bound     default/scbe-accept-vol1                    gold                        1m
```
### Running a Pod with a Ubiquity volume
The creation of a Pod/Deployment will cause the Ubiquity to: 
* Attach the volume to the host
* Rescan and discover the multipath device of the new volume
* Create xfs or ext4 filesystem on the device
* Mount the new multipath device on /ubiquity/[WWN of the volume]

Kubernetes create Pod template:
```bash
#> kubectl create -f <path/to/pod.yml>
```

For example, to start a Pod `acceptance-pod-test` with the created PVC `scbe-accept-vol1`:

```bash
#> kubectl create -f deploy/scbe_volume_with_pod.yml
```

You can display the new created pod. In addition you can write data into the  persistent volume on IBM FlashSystem A9000R.
```bash
#> kubectl get pod
NAME                 READY      STATUS    RESTARTS   AGE
acceptance-pod-test   1/1       Running   0          1m

#> kubectl exec acceptance-pod-test df | egrep "/mnt|^Filesystem"
Filesystem           1K-blocks      Used Available Use% Mounted on
/dev/mapper/mpathacg   9755384     32928   9722456   0% /mnt

#> kubectl exec acceptance-pod-test mount | egrep "/mnt"
/dev/mapper/mpathacg on /data type xfs (rw,seclabel,relatime,attr2,inode64,noquota)

#> kubectl exec acceptance-pod-test touch /mnt/FILE
#> kubectl exec acceptance-pod-test ls /mnt/FILE
File
```

Now you can also display the newly attached volume on the host.
```bash
#> multipath -ll
mpathacg (36001738cfc9035eb0000000000cbb306) dm-8 IBM     ,2810XIV         
size=9.3G features='1 queue_if_no_path' hwhandler='0' wp=rw
`-+- policy='service-time 0' prio=1 status=active
  |- 8:0:0:1 sdb 8:16 active ready running
  `- 9:0:0:1 sdc 8:32 active ready running

#> mount |grep ubiquity
/dev/mapper/mpathacg on /ubiquity/6001738CFC9035EB0000000000CBB306 type xfs (rw,relatime,seclabel,attr2,inode64,noquota)

#> df | egrep "ubiquity|^Filesystem" 
Filesystem                       1K-blocks    Used Available Use% Mounted on
/dev/mapper/mpathacg               9755384   32928   9722456   1% /ubiquity/6001738CFC9035EB0000000000CFF306

```

### Deleting a Pod  with a volume
The kuberenetes delete command will cause the Ubiquity to detach the volume from the host.
```bash
#> kubectl delete acceptance-pod-test
```

### Removing a volume
In order to remove the volume, we need to remove its PVC:
```bash
#> kubectl delete pvc scbe-accept-vol1
persistentvolumeclaim "scbe-accept-vol1" deleted

#> kubectl get pvc,pv
NAME                             STATUS    VOLUME                                      CAPACITY   ACCESSMODES   STORAGECLASS             AGE
No resources found.

NAME                                          CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS     CLAIM                                STORAGECLASS             REASON    AGE
No resources found.
```
