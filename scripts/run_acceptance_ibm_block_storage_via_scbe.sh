#!/usr/bin/env bash

set -ex

scripts=$(dirname $0)

echo "Creating Storage class...."
kubectl create -f $scripts/../deploy/scbe_volume_storage_class.yml

echo "Listing Storage classes"
kubectl get storageclass


echo "Creating Persistent Volume Claim..."
kubectl create -f $scripts/../deploy/scbe_volume_pvc.yml


echo "Listing Persistent Volume Claim..."
kubectl get pvc


echo "Listing Persistent Volume..."
kubectl get pv


echo "Creating Test Pod"
kubectl create -f $scripts/../deploy/scbe_volume_with_pod.yml
sleep 10

echo "Listing pods"
kubectl get pods

echo "Writing success.txt to mounted volume"
kubectl exec acceptance-pod-test -c acceptance-pod-test-con touch /mnt/success.txt

echo "Reading from mounted volume"
kubectl exec acceptance-pod-test -c acceptance-pod-test-con ls /mnt


echo "Cleaning test environment"

echo "Deleting Pod"
kubectl delete -f $scripts/../deploy/scbe_volume_with_pod.yml

echo "Deleting Persistent Volume Claim"
kubectl delete -f $scripts/../deploy/scbe_volume_pvc.yml

echo "Listing PVC"
kubectl get pvc

echo "Listing PV"
kubectl get pv

echo "Deleting Storage Class"
kubectl delete -f $scripts/../deploy/scbe_volume_storage_class.yml

echo "Listing Storage Classes"
kubectl get storageclass
