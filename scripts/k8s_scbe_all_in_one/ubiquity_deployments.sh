#!/bin/bash -e

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
# The script start\stop ubiquity deployments with the right order.
# The right order is to start ubiquity and provisioner before ubiquity-db and vise versa for stop.
# The script assume that you already installed ubiquity via the ubiquity_install.sh script.
#
# Start flow order:
#   1. Validate ubiquity pvc exist (exit if not)
#   2. Create ubiquity deployment
#   3. Create ubiquity-k8s-provisioner deployment
#   4. Create ubiquity-k8s-flex Daemonset
#   4. Create ubiquity-db-deployment deployment
#
# Start flow order:
#   1. Delete ubiquity-db-deployment deployment
#   2. Delete Ubiquity-k8s-provisioner deployment
#   3. Create ubiquity-k8s-flex Daemonset
#   4. Delete Ubiquity deployment
#
# Assuming the following components was already created during the ubiquity_install.sh script.
#   - Ubiquity service
#   - Ubiquity-db service
#   - configmap k8s-config for Ubiquity-k8s-provisioner
#   - Storage class that match for the UbiquityDB PVC
#   - PVC for the UbiquityDB
#   - FlexVolume already installed on each minion and configured well
#
# Note : More details see usage below.
# -------------------------------------------------------------------------


function usage()
{
   echo "Usage, $0 [start|stop|status|getall|getallwide|help]"
   echo "    Options"
   echo "       start  : Create ubiquity, provisioner deployments, flex daemonset and ubiquity-db deployments"
   echo "       stop   : Delete ubiquity-db(wait for deletion), provisioner deployment, flex daemonset, ubiquity deployment"
   echo "       status : kubectl get to all the ubiquity components"
   echo "       getall : kubectl get configmap,storageclass,pv,pvc,service,daemonset,deployment,pod"
   echo "       getallwide : getall but with -o wide"
   echo "       help      : Show this usage"
   exit 1
}

function help()
{
  usage
}

function start()
{
    echo "Make sure ${UBIQUITY_DB_PVC_NAME} exist and bounded to PV (if not exit), before starting."
    wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 5 2

    kubectl create -f ${YML_DIR}/ubiquity-deployment.yml
    kubectl create -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
    kubectl create -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl create -f ${YML_DIR}/ubiquity-db-deployment.yml
    echo "Finished to start ubiquity components. Run $0 status to get more details."
}

function stop()
{
    # TODO delete the deployments only if its actually exist
    kubectl delete -f $YML_DIR/ubiquity-db-deployment.yml
    echo "Wait for ubiquity-db deployment deletion..."
    wait_for_item_to_delete deployment ubiquity-db 10 4
    wait_for_item_to_delete pod "ubiquity-db-" 10 4 regex  # to match the prefix of the pod

    kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
    kubectl delete -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl delete -f $YML_DIR/ubiquity-deployment.yml
    echo "Finished to stop ubiquity deployments. Run $0 status to get more details."
}


function status()
{
    # kubectl get on all the ubiquity components, if one of the components are not found
    rc=0
    kubectl get configmap k8s-config || rc=$?
    echo ""
    kubectl get storageclass | egrep "ubiquity|^NAME"  || rc=$?
    echo ""
    kubectl get pv/ibm-ubiquity-db pvc/ibm-ubiquity-db svc/ubiquity svc/ubiquity-db  daemonset/ubiquity-k8s-flex deploy/ubiquity deploy/ubiquity-db deploy/ubiquity-k8s-provisioner  || rc=$?
    echo ""
    kubectl get pod | egrep "^ubiquity|^NAME" || rc=$?

    if [ $rc != 0 ]; then
       echo ""
       echo "Ubiquity status [NOT ok]. Some components are missing(review the output above)"
       exit 5
    #else
    #   # TODO verify deployment status
    #   verify_deployments_status ubiquity ubiquity-db ubiquity-k8s-provisioner
    fi
}

function verify_deployments_status()
{
       deployments="$@"
       bad_deployment=""
       for deployment_name in $deployments; do
            if ! is_deployment_ok ${deployment_name}; then
               echo "Deployment [${deployment_name}] status not OK"
               bad_deployment="$deployment_name $bad_deployment"
            fi
       done
       if [ -n "$bad_deployment" ]; then
           echo ""
           echo "Ubiquity status [NOT ok]. Some deployments are NOT ok (review the output above)."
           exit 6
       fi
       # TODO also need to validate that the daemon set is in the current state

       echo "Ubiquity status [OK]"
       exit 0
}
function getall()
{
    flag="$1"
    kubectl get $flag configmap,storageclass,pv,pvc,service,daemonset,deployment,pod || :
}

function getallwide()
{
    getall "-o wide"
}



scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
FLEX_DIRECTORY='/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex'

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl > /dev/null 2>&1 || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
. $UTILS # include utils for wait function and status
[ $# -ne 1 ] && usage
action=$1
[ $action != "start" -a $action != "stop" -a $action != "status" -a $action != "getall" -a $action != "getallwide" -a $action != "help" ] && usage


$1


