#!/bin/bash -x

YML_DIR="./yamls"
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-database
UTILS=../$scripts/acceptance_utils.sh

[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }

. $UTILS # include utils for wait function and status

pvname=`kubectl get pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
kubectl delete -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for PVC ${UBIQUITY_DB_PVC_NAME} and PV $pvname to be deleted, before delete ubiquity and provisioner."
wait_for_item_to_delete pvc ${UBIQUITY_DB_PVC_NAME} 10 2
wait_for_item_to_delete pv $pvname 10 2

kubectl delete -f ${YML_DIR}/storage-class.yml
kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
kubectl delete configmap k8s-config
kubectl delete -f $YML_DIR/ubiquity-deployment.yml
kubectl delete -f $YML_DIR/ubiquity-service.yml
kubectl delete -f $YML_DIR/ubiquity-db-service.yml

echo "Finished to remove all ubiquity components. (Attention : You should remove flex from all the minion manually.)"


