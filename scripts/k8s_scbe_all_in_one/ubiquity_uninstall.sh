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


function usage()
{
  cmd=`basename $0`
  echo "USAGE   $cmd [-n <namespace>] [-h]"
  echo "  Options:"
  echo "    -n [<namespace>] : Namespace of ubiquity. By default its \"ubiquity\" namespace"
  echo "    -h : Display this usage"
  exit 1
}

# Variables
scripts=$(dirname $0)
YML_DIR="./yamls"
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
UTILS=$scripts/ubiquity_utils.sh
flex_conf="ubiquity-k8s-flex.conf"
K8S_CONFIGMAP_FOR_PROVISIONER=k8s-config
FLEX_K8S_DIR=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex

# Handle flags
NS="ubiquity" # default namespace
while getopts ":n:" opt; do
  case $opt in
    n)
      NS=$OPTARG
      ;;
    h)
      usage
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      usage
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      usage
      ;;
  esac
done
nsf="--namespace ${NS}" # namespace flag for kubectl command

kubectl_delete="kubectl delete $nsf --ignore-not-found=true"
echo "Uninstall Ubiquity from namespace [$NS]..."

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
kubectl get namespace $NS >/dev/null 2>&1 || { echo "[$NS] namespace not exist. Stop the uninstall process."; exit 3; }

. $UTILS # include utils for wait function and status

# First phase : delete the ubiquity-db deployment and ubiquity-db-pvc before deleting ubiquity and provisioner.
if kubectl get $nsf deployment ubiquity-db >/dev/null 2>&1; then
    $kubectl_delete -f $YML_DIR/ubiquity-db-deployment.yml
    echo "Wait for ubiquity-db deployment deletion..."
    wait_for_item_to_delete deployment ubiquity-db 10 4 "" $NS
    wait_for_item_to_delete pod "ubiquity-db-" 10 4 regex $NS # to match the prefix of the pod
fi
pvname=`kubectl get $nsf pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
$kubectl_delete -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for PVC ${UBIQUITY_DB_PVC_NAME} and PV $pvname to be deleted, before delete ubiquity and provisioner."
wait_for_item_to_delete pvc ${UBIQUITY_DB_PVC_NAME} 10 3 "" $NS
[ -n "$pvname" ] && wait_for_item_to_delete pv $pvname 10 3 "" $NS



# Second phase : Delete all the stateless components
$kubectl_delete -f ${YML_DIR}/storage-class.yml
$kubectl_delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
$kubectl_delete configmap ${K8S_CONFIGMAP_FOR_PROVISIONER}

$kubectl_delete -f $YML_DIR/ubiquity-k8s-flex-daemonset.yml
$kubectl_delete configmap $flex_conf

$kubectl_delete -f $YML_DIR/ubiquity-deployment.yml
$kubectl_delete -f $YML_DIR/ubiquity-service.yml
$kubectl_delete -f $YML_DIR/ubiquity-db-service.yml
$kubectl_delete -f $YML_DIR/ubiquity-namespace.yml


echo "Ubiquity uninstall finished."

