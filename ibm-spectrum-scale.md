# IBM Spectrum Scale

IBM Spectrum Scale can be used as persistent storage for Kubernetes via Ubiquity service. Ubiquity communicates with the IBM Spectrum Scale through IBM Spectrum Scale management api version 2. Filesets are created on IBM Spectrum Scale and made available for Ubiquity FlexVolume and Ubiquity Dynamic Provisioner.

# Usage examples for Ubiquity Dynamic Provisioner and FlexVolume
The IBM official solution for Kubernetes, based on the Ubiquity project, is referred to as IBM Storage Enabler for Containers. You can download the installation package and its documentation (including full usage examples) from [IBM Fix Central](https://www.ibm.com/support/fixcentral/swg/selectFixes?parent=Software%2Bdefined%2Bstorage&product=ibm/StorageSoftware/IBM+Spectrum+Connect&release=All&platform=Linux&function=all).

Usage examples index:
* [Example 1 : Basic flow for running a stateful container in a pod](#example-1--basic-flow-for-running-a-stateful-container-with-ubiquity-volume)
* [Example 2 : Basic flow breakdown](#example-2--basic-flow-breakdown)
* [Example 3 : Deployment fail over](#example-3--deployment-fail-over-example)

## Example 1 : Basic flow for running a stateful container with Ubiquity volume
Flow overview:
1. Create a StorageClass `spectrumscale-primaryfs` that refers to Spectrum Scale filesysten `primaryfs`.
2. Create a PVC `pvc1` that uses the StorageClass `spectrumscale-primaryfs`.
3. Create a Pod `pod1` with container `container1` that uses PVC `pvc1`.
3. Start I/Os into `/data/myDATA` in `pod1\container1`.
4. Delete the `pod1` and then create a new `pod1` with the same PVC and verify that the file `/data/myDATA` still exists.
5. Delete the `pod1` `pvc1`, `pv` and storage class `spectrumscale-primaryfs`.

Relevant yml files (`storage-class-primaryfs.yml`, `pvc1.yml` and `pod1.yml`):
```bash
#> cat storage-class-primaryfs.yml pvc1.yml pod1.yml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: "spectrumscale-primaryfs"
  labels:
    product: ibm-storage-enabler-for-containers
provisioner: "ubiquity/flex"
parameters:
  backend: "spectrum-scale"
  filesystem: "primaryfs"
  type: fileset


kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: "pvc1"
  annotations:
    volume.beta.kubernetes.io/storage-class: "spectrumscale-primaryfs"
spec:
  accessModes:
    - ReadWriteOnce # Ubiquity Spectrum Scale backend supports ReadWriteOnce and ReadWriteMany mode.
  resources:
    requests:
      storage: 1Gi  # Size in Gi unit only


kind: Pod
apiVersion: v1
metadata:
  name: pod1          # Pod name
spec:
  containers:
  - name: container1  # Container name
    image: alpine:latest
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "while true; do sleep 30; done;" ]
    volumeMounts:
      - name: vol1
        mountPath: "/data"  # Where to mount the vol1(pvc1)
  restartPolicy: "Never"
  volumes:
    - name: vol1
      persistentVolumeClaim:
        claimName: pvc1
```

Running the basic flow:
```bash
#> kubectl create -f storage-class-primaryfs.yml -f pvc1.yml -f pod1.yml
storageclass "spectrumscale-primaryfs" created
persistentvolumeclaim "pvc1" created
pod "pod1" created

#### Wait for PV to be created and pod1 to be in the Running state...

#> kubectl get storageclass spectrumscale-primaryfs
NAME                      PROVISIONER     AGE
spectrumscale-primaryfs   ubiquity/flex   1m

#> kubectl get pvc pvc1
NAME      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS              AGE
pvc1      Bound     pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24   1Gi        RWO            spectrumscale-primaryfs   1m

#> kubectl get pv pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM          STORAGECLASS              REASON    AGE
pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24   1Gi        RWO            Delete           Bound     default/pvc1   spectrumscale-primaryfs             2m

#> kubectl get pod pod1
NAME      READY     STATUS    RESTARTS   AGE
pod1      1/1       Running   0          2m

#> kubectl exec pod1  -c container1 -- sh -c "dd if=/dev/zero of=/data/myDATA bs=10M count=1"

#> kubectl exec pod1  -c container1 -- sh -c "ls -l /data/myDATA"
-rw-r--r--    1 root     root      10485760 Nov 25 14:24 /data/myDATA

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
pod1      1/1       Running   0          46s

#### Verify the /data/myDATA still exist

#> kubectl exec pod1  -c container1 -- sh -c "ls -l /data/myDATA"
-rw-r--r--    1 root     root      10485760 Nov 25 14:24 /data/myDATA

### Delete pod1, pvc1, pv and the gold storage class
#> kubectl delete -f pod1.yml -f pvc1.yml -f storage-class-primaryfs.yml
pod "pod1" deleted
persistentvolumeclaim "pvc1" deleted
storageclass "spectrumscale-primaryfs" deleted
```


## Example 2 : Basic flow breakdown
This section describes separate steps of the generic flow in greater detail.


### Creating a Storage Class
For example, to create a Storage Class named `spectrumscale-primaryfs` that refers to a Spectrum Scale filesystem `primaryfs` with fileset quota enabled. As a result, every volume from this storage class will be provisioned on the Spectrum  Scale and each volume is a fileset in a Spectrum Scale filesystem.
```bash
#> cat storage-class-primaryfs.yml
kind: StorageClass
apiVersion: storage.k8s.io/v1beta1
metadata:
  name: "spectrumscale-primaryfs"
  labels:
    product: ibm-storage-enabler-for-containers
provisioner: "ubiquity/flex"
parameters:
  backend: "spectrum-scale"
  filesystem: "primaryfs"
  type: fileset

#> kubectl create -f storage-class-primaryfs.yml
storageclass "spectrumscale-primaryfs" created
```

List the newly created Storage Class:
```bash
#> kubectl get storageclass spectrumscale-primaryfs
NAME                      PROVISIONER     AGE
spectrumscale-primaryfs   ubiquity/flex   1m
```

### Creating a PersistentVolumeClaim
To create a PVC `pvc1` with size `1Gi` that uses the `spectrumscale-primaryfs` Storage Class:
```bash
#> cat pvc1.yml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: "pvc1"
  annotations:
    volume.beta.kubernetes.io/storage-class: "spectrumscale-primaryfs"
spec:
  accessModes:
    - ReadWriteOnce # Ubiquity Spectrum Scale backend supports ReadWriteOnce and ReadWriteMany mode.
  resources:
    requests:
      storage: 1Gi  # Size in Gi unit only

#> kubectl create -f pvc1.yml
persistentvolumeclaim "pvc1 created
```

Ubiquity Dynamic Provisioner automatically creates a PersistentVolume (PV) and binds it to the PVC. The PV name will be PVC-ID. The volume name on the storage will be `[PVC-ID]`.

List a PersistentVolumeClaim and PersistentVolume
```bash
#> kubectl get pvc pvc1
NAME      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS              AGE
pvc1      Bound     pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24   1Gi        RWO            spectrumscale-primaryfs   1m

#> kubectl get pv pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM          STORAGECLASS              REASON    AGE
pvc-2073e8fd-f0bd-11e8-a8f1-000c29e45a24   1Gi        RWO            Delete           Bound     default/pvc1   spectrumscale-primaryfs             2m
```

### Create a Pod with an Ubiquity volume
The creation of a Pod/Deployment causes the FlexVolume to:
* Create a symbolic link /var/lib/kubelet/pods/[POD-ID]/volumes/ibm~ubiquity-k8s-flex/[PVC-ID] -> [fileset linkpath]

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
    image: alpine:latest
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "while true; do sleep 30; done;" ]
    volumeMounts:
      - name: vol1
        mountPath: "/data"  # Where to mount the vol1(pvc1)
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

#> kubectl exec pod1 -c container1  -- sh -c "df -h /data"
Filesystem                Size      Used Available Use% Mounted on
primaryfs               100.0G      4.5G     95.5G   4% /data

#> kubectl exec pod1  -c container1 -- sh -c "dd if=/dev/zero of=/data/myDATA bs=10M count=1"

#> kubectl exec pod1  -c container1 -- sh -c "ls -l /data/myDATA"
-rw-r--r--    1 root     root      10485760 Nov 25 14:24 /data/myDATA
```

### Deleting a Pod
The Kuberenetes delete Pod command:
* Removes symbolic link /var/lib/kubelet/pods/[POD-ID]/volumes/ibm~ubiquity-k8s-flex/[PVC-ID] -> [fileset link path]

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
#> kubectl delete -f storage-class-primaryfs.yml
storageclass "spectrumscale-primaryfs" deleted
```


## Example 3 : Deployment fail over example
This section describes how to run stateful Pod with k8s Deployment object, and then delete the Pod and see how kubernetes schedule the pod on different node and the PV follows.


- Prerequisits
1. Create the same storage class (as previous example)
```bash
#> kubectl create -f storage-class-primaryfs.yml
storageclass "spectrumscale-primaryfs" created
```
2. Create the PVC (as previous example)
```bash
#> kubectl create -f pvc1.yml
persistentvolumeclaim "pvc1" created
```

- Create Kubernetes Deployment with stateful POD (on node6) and write some data inside
```bash
#> cat deployment1.yml
apiVersion: "extensions/v1beta1"
kind: Deployment
metadata:
  name: deployment1
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: deployment1
    spec:
      containers:
      - name: container1
        image: alpine:latest
        command: [ "/bin/sh", "-c", "--" ]
        args: [ "while true; do sleep 30; done;" ]
        volumeMounts:
          - name: pvc
            mountPath: "/data"
      volumes:
      - name: pvc
        persistentVolumeClaim:
          claimName: pvc1

#> kubectl create -f deployment1.yml
deployment "deployment1" created

#> kubectl get -o wide deploy,pod
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE       CONTAINERS   IMAGES          SELECTOR
deploy/deployment1   1         1         1            1           1m        container1   alpine:latest   app=deployment1

NAME                             READY     STATUS    RESTARTS   AGE       IP            NODE
po/deployment1-df6dd77d4-6kb7k   1/1       Running   0          1m        10.244.3.48   node6

#> pod=`kubectl get pod | awk '/deployment1/{print $1}'`
#> echo $pod
deployment1-df6dd77d4-6kb7k

#> kubectl exec $pod -- /bin/sh -c "echo COOL > /data/file"
#> kubectl exec $pod -- /bin/sh -c "cat /data/file"
COOL
```

- Delete the POD so Kubernetes will reschedule the POD on a diffrent node (node5)
```bash
#> kubectl delete pod $pod
pod "deployment1-df6dd77d4-6kb7k" deleted
#> kubectl get -o wide deploy,pod
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE       CONTAINERS   IMAGES          SELECTOR
deploy/deployment1   1         1         1            0           4m        container1   alpine:latest   app=deployment1

NAME                             READY     STATUS              RESTARTS   AGE       IP            NODE
po/deployment1-df6dd77d4-6kb7k   1/1       Terminating         0          4m        10.244.3.48   node6
po/deployment1-df6dd77d4-ngfdp   0/1       ContainerCreating   0          23s       <none>        node5

#############
## Wait a few seconds
#############

#> kubectl get -o wide deploy,pod
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE       CONTAINERS   IMAGES          SELECTOR
deploy/deployment1   1         1         1            1           5m        container1   alpine:latest   app=deployment1

NAME                             READY     STATUS    RESTARTS   AGE       IP            NODE
po/deployment1-df6dd77d4-ngfdp   1/1       Running   0          1m        10.244.2.47   node5

#############
## Now check data remains
#############
#> pod=`kubectl get pod | awk '/deployment1/{print $1}'`
#> echo $pod
deployment1-df6dd77d4-ngfdp
#> kubectl exec $pod -- /bin/sh -c "cat /data/file"
COOL

```

- Tier down the Deployment, PVC, PV and Storage Class
```bash
#> kubectl delete -f deployment1.yml -f pvc1.yml -f storage-class-primaryfs.yml
deployment "deployment1" deleted
persistentvolumeclaim "pvc1" deleted
storageclass "spectrumscale-primaryfs" deleted
```

</br>

**Note:** For detailed usage examples, refer to the IBM Storage Enabler for Containers [user guide](https://www-945.ibm.com/support/fixcentral/swg/selectFixes?parent=Software%2Bdefined%2Bstorage&product=ibm/StorageSoftware/IBM+Spectrum+Connect&release=All&platform=Linux&function=all).
