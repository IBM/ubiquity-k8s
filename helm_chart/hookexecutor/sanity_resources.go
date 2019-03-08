package hookexecutor

/**
 * The file is used to define the resources that will be created during a sanity
 * test. All the reosurces are in yaml format, but stored in the file as separate
 * strings. In this way, we can package these resources at compile-time, and use
 * them at any time without touching local files in the container.
 */

var sanityPod = `
kind: Pod
apiVersion: v1
metadata:
  name: sanity-pod
  namespace: default
spec:
  containers:
  - name: container1
    image: alpine:3.8
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "while true; do sleep 30; done;" ]
    volumeMounts:
      - name: vol1
        mountPath: "/data"
  restartPolicy: "Never"
  volumes:
    - name: vol1
      persistentVolumeClaim:
        claimName: sanity-pvc
`

var sanityPvc = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sanity-pvc
  namespace: default
  labels:
    pv-name: sanity-pv
  annotations:
    volume.beta.kubernetes.io/storage-class: ""
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
`
