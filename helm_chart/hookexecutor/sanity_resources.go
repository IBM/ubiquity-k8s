package hookexecutor

var sanityPod = `
kind: Pod
apiVersion: v1
metadata:
  name: sanity-pod
  namespace: default
spec:
  containers:
  - name: container1
    image: alpine:latest
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
  annotations:
    volume.beta.kubernetes.io/storage-class: ""
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
`
