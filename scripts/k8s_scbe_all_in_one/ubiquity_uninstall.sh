#!/bin/bash

###################################################
# Copyright 2017 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
###################################################

# -------------------------------------------------------------------------
# The script uninstall all Ubiquity components (including the ubiquity data which locate in PV ibm-ubiquity-db).
#
# Delete the following components by order:
#   1. ubiquity-db deployment (wait for deletion)
#   2. PVC for the UbiquityDB (wait for deletion of PVC and PV)
#   3. Storage class that match for the UbiquityDB PVC
#   4. Ubiquity-k8s-provisioner deployment
#   5. k8s-config configmap
#   6. ubiquity-k8s-flex daemonset
#   7. $flex_conf configmap
#   8. Ubiquity deployment
#   9. Ubiquity service
#   10.Ubiquity-db service
#
# Note : The script delete the flex daemon set but it does NOT delete the flex driver on the nodes.
#        Its not mandatory to delete the flex driver.
#        If user wants to delete the flex manually, then here are the steps to run on each node:
#           1. remove the flex directory : /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
#           2. restart the kubelet service
#
# Prerequisites to run this test:
#   - The script assumes all the yamls exist under ./yamls and updated with relevant configuration.
#   - Run the script in the Kubernetes master node where you have access to kubectl command.
#   - See usage function below for more details about flags.
# -------------------------------------------------------------------------


# Variables
scripts=$(dirname $0)
YML_DIR="./yamls"
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
UTILS=$scripts/ubiquity_utils.sh
flex_conf="ubiquity-k8s-flex.conf"
K8S_CONFIGMAP_FOR_PROVISIONER=k8s-config
FLEX_K8S_DIR=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }

. $UTILS # include utils for wait function and status

# First phase : delete the ubiquity-db deployment and ubiquity-db-pvc before deleting ubiquity and provisioner.
if kubectl get deployment ubiquity-db >/dev/null 2>&1; then
    kubectl delete -f $YML_DIR/ubiquity-db-deployment.yml
    echo "Wait for ubiquity-db deployment deletion..."
    wait_for_item_to_delete deployment ubiquity-db 10 4
    wait_for_item_to_delete pod "ubiquity-db-" 10 4 regex  # to match the prefix of the pod
fi
pvname=`kubectl get pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
kubectl delete -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for PVC ${UBIQUITY_DB_PVC_NAME} and PV $pvname to be deleted, before delete ubiquity and provisioner."
wait_for_item_to_delete pvc ${UBIQUITY_DB_PVC_NAME} 10 3
wait_for_item_to_delete pv $pvname 10 3



# Second phase : Delete all the stateless components
kubectl delete -f ${YML_DIR}/storage-class.yml
kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
kubectl delete configmap ${K8S_CONFIGMAP_FOR_PROVISIONER}

kubectl delete -f $YML_DIR/ubiquity-k8s-flex-daemonset.yml
kubectl delete configmap $flex_conf

kubectl delete -f $YML_DIR/ubiquity-deployment.yml
kubectl delete -f $YML_DIR/ubiquity-service.yml
kubectl delete -f $YML_DIR/ubiquity-db-service.yml

echo "Ubiquity uninstall finished."

