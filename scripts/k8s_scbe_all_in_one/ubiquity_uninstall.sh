#!/bin/bash -x

scripts=$(dirname $0)
YML_DIR="./yamls"
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
UTILS=$scripts/ubiquity_utils.sh
flex_conf="ubiquity-k8s-flex.conf"

[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }

. $UTILS # include utils for wait function and status

# First phase : MUST to delete the ubiquity-db deployment and ubiquity-db-pvc before deleting ubiquity and provisioner.
if kubectl get deployment ubiquity-db >/dev/null 2>&1; then
    kubectl delete -f $YML_DIR/ubiquity-db-deployment.yml
    sleep 30 # TODO wait till deployment stopped including the POD.
fi
pvname=`kubectl get pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
kubectl delete -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for PVC ${UBIQUITY_DB_PVC_NAME} and PV $pvname to be deleted, before delete ubiquity and provisioner."
wait_for_item_to_delete pvc ${UBIQUITY_DB_PVC_NAME} 10 3
wait_for_item_to_delete pv $pvname 10 3

# Second phase : Delete all the stateless components
kubectl delete -f ${YML_DIR}/storage-class.yml
kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
kubectl delete configmap k8s-config

kubectl delete -f $YML_DIR/ubiquity-k8s-flex-daemonset.yml
kubectl delete configmap $flex_conf

kubectl delete -f $YML_DIR/ubiquity-deployment.yml
kubectl delete -f $YML_DIR/ubiquity-service.yml
kubectl delete -f $YML_DIR/ubiquity-db-service.yml

echo "Finished to remove all ubiquity components. (Attention :  removal of ubiquity FlexVolume is a manual operation need to be done on each minion"

# TODO make the script with -e and delete component only if there are not already exist.

