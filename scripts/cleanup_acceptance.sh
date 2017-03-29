#!/usr/bin/env bash

set -e

scripts=$(dirname $0)


echo "Cleaning test environment"

echo "Deleting Pod"
kubectl delete -f $scripts/../deploy/pod.yml

echo "Deleting Persistent Volume Claim"
kubectl delete -f $scripts/../deploy/pvc_fileset.yml

echo "Listing PVC"
kubectl get pvc

echo "Listing PV"
kubectl get pv

echo "Deleting Storage Class"
kubectl delete -f $scripts/../deploy/storage_class_fileset.yml

echo "Listing Storage Classes"
kubectl get storageclass