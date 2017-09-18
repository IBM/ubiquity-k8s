#!/bin/bash -x

YML_DIR="./yamls"

[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }

kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
kubectl delete configmap k8s-config
kubectl delete -f $YML_DIR/ubiquity-deployment.yml
kubectl delete -f $YML_DIR/ubiquity-service.yml
kubectl delete -f $YML_DIR/ubiquity-db-service.yml



