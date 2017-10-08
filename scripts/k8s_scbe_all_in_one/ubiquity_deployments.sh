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
   echo "Usage, $0 -a [`echo $actions | sed 's/ /|/g'`] [-n <namespace>] [-h]"
   echo " -a <action>"
   echo "   start  : Create ubiquity, provisioner deployments, flex daemonset and ubiquity-db deployments"
   echo "   stop   : Delete ubiquity-db(wait for deletion), provisioner deployment, flex daemonset, ubiquity deployment"
   echo "   status : kubectl get to all the ubiquity components"
   echo "   getall : kubectl get configmap,storageclass,pv,pvc,service,daemonset,deployment,pod"
   echo "   getallwide : getall but with -o wide"
   echo "   collect_logs : Create a directory with all Ubiquity logs"
   echo "   sanity    : create and delete pvc and pod"
   echo " -n <namespace>  : Optional, by default its \"ubiquity\""
   echo " -h : Display this usage"

   exit 1
}

function help()
{
  usage
}

function start()
{
    echo "Make sure ${UBIQUITY_DB_PVC_NAME} exist and bounded to PV (if not exit), before starting."
    wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 5 2 $NS

    kubectl create $nsf -f ${YML_DIR}/ubiquity-deployment.yml
    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl create $nsf -f ${YML_DIR}/ubiquity-db-deployment.yml
    # TODO add some wait for deployment before run the ubiquity-db
    echo "Finished to start ubiquity components. Note : View ubiquity status by : $> $0 -a status -n $NS"
}

function stop()
{
    kubectl_delete="kubectl delete $nsf --ignore-not-found=true"

    # TODO In addition it can works on the k8s object instead on yaml files
    $kubectl_delete -f $YML_DIR/ubiquity-db-deployment.yml
    echo "Wait for ubiquity-db deployment deletion..."
    wait_for_item_to_delete deployment ubiquity-db 10 4 "" $NS
    wait_for_item_to_delete pod "ubiquity-db-" 10 4 regex $NS  # to match the prefix of the pod

    $kubectl_delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
    $kubectl_delete -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    $kubectl_delete -f $YML_DIR/ubiquity-deployment.yml
    echo "Finished to stop ubiquity deployments.   Note : View ubiquity status by : $> $0 -a status -n $NS"
}

function collect_logs()
{
    # Get logs from all ubiquity deployments and pods into a directory

    time=`date +"%m-%d-%Y-%T"`
    logdir=./ubiquity_collect_logs_$time
    klog="kubectl logs $nsf"
    mkdir $logdir
    ubiquity_log_name=${logdir}/ubiquity.log
    ubiquity_db_log_name=${logdir}/ubiquity-db.log
    ubiquity_provisioner_log_name=${logdir}/ubiquity-k8s-provisioner.log
    ubiquity_status_log_name=ubiquity_deployments_status.log

    # kubectl logs on all deployments
    echo "$klog deploy/ubiquity"
    $klog deploy/ubiquity > ${ubiquity_log_name} 2>&1 || :
    echo "$klog deploy/ubiquity-db"
    $klog deploy/ubiquity-db > ${ubiquity_db_log_name} 2>&1 || :
    echo "$klog deploy/ubiquity-k8s-provisioner"
    $klog deploy/ubiquity-k8s-provisioner > ${ubiquity_provisioner_log_name} 2>&1 || :
    files_to_collect="$ubiquity_log_name ${ubiquity_db_log_name} ${ubiquity_provisioner_log_name}"

    # kubectl logs on flex PODs
    for flex_pod in `kubectl get $nsf pod | grep ubiquity-k8s-flex | awk '{print $1}'`; do
       echo "$klog pod ${flex_pod}"
       $klog pod ${flex_pod} > ${logdir}/${flex_pod}.log 2>&1 || :
       files_to_collect="${files_to_collect} ${logdir}/${flex_pod}.log"
    done
    echo "$0 status"
    status > ${logdir}/${ubiquity_status_log_name} 2<&1 || :

    echo ""
    echo "Finish collecting Ubiquity logs inside directory -> $logdir"
}


function status()
{
    # kubectl get on all the ubiquity components, if one of the components are not found
    rc=0

    cmd="kubectl get storageclass | egrep \"ubiquity|^NAME\""
    echo $cmd
    echo '---------------------------------------------------------------------'
    kubectl get storageclass | egrep "ubiquity|^NAME" || rc=$?
    echo ""

    cmd="kubectl get $nsf cm/k8s-config pv/ibm-ubiquity-db pvc/ibm-ubiquity-db svc/ubiquity svc/ubiquity-db  daemonset/ubiquity-k8s-flex deploy/ubiquity deploy/ubiquity-db deploy/ubiquity-k8s-provisioner"
    echo $cmd
    echo '---------------------------------------------------------------------'
    $cmd  || rc=$?

    echo ""
    cmd="kubectl get $nsf pod | egrep \"^ubiquity|^NAME\""
    echo $cmd
    echo '---------------------------------------------------------------------'
    kubectl get $nsf pod | egrep "^ubiquity|^NAME" || rc=$?

    if [ $rc != 0 ]; then
       echo ""
       echo "Ubiquity status [NOT ok]. Some components are missing(review the output above)"
       exit 5
    #else
    #   # TODO verify deployment status
    #   verify_deployments_status ubiquity ubiquity-db ubiquity-k8s-provisioner $NS
    fi
}

function verify_deployments_status()
{
       # TODO add notion of namespace
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
    # Args : $1 additional flags to kubectl like -o wide
    cmd="kubectl get $1 namespace,pv,storageclass"
    echo $cmd
    $cmd || :


    echo ""
    echo "Default namespace:"
    echo "=================="
    cmd="kubectl get --namespace default $1 configmap,pvc,service,daemonset,deployment,pod"
    echo $cmd
    $cmd || :

    [ "$NS" == "default" ] && return
    echo ""
    echo "$NS namespace:"
    echo "=================="
    cmd="kubectl get --namespace $NS $1 configmap,pvc,service,daemonset,deployment,pod"
    echo $cmd
    $cmd || :
}

function getallwide()
{
    getall "-o wide"
}

function sanity()
{
    pvc="sanity-pvc"
    pod="sanity-pod"

    echo "--------------------------------------------------------------"
    echo "Sanity description:"
    echo "    1. Create $pvc, $pod and wait for creation."
    echo "    2. Delete the $pod, $pvc and wait for deletion."
    echo "    Note : Uses yamls from directory ${SANITY_YML_DIR}, and uses the ubiquity storage class."
    echo "--------------------------------------------------------------"
    echo ""

    kubectl create $nsf -f ${SANITY_YML_DIR}/${pvc}.yml
    wait_for_item pvc ${pvc} ${PVC_GOOD_STATUS} 10 3 $NS
    pvname=`kubectl get $nsf pvc ${pvc} --no-headers -o custom-columns=name:spec.volumeName`

    kubectl create $nsf -f ${SANITY_YML_DIR}/${pod}.yml
    wait_for_item pod ${pod} Running 100 3 $NS

    kubectl delete $nsf -f ${SANITY_YML_DIR}/${pod}.yml
    wait_for_item_to_delete pod ${pod} 100 3 "" $NS
    kubectl delete $nsf -f ${SANITY_YML_DIR}/${pvc}.yml
    wait_for_item_to_delete pvc ${pvc} 10 2 "" $NS
    wait_for_item_to_delete pv $pvname 10 2 "" $NS

    echo ""
    echo "Ubiquity sanity finished successfully."
}


scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
SANITY_YML_DIR="$scripts/yamls/sanity_yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
FLEX_DIRECTORY='/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex'
actions="start stop status getall getallwide collect_logs sanity"

# Handle flags
NS="ubiquity" # default namespace
while getopts "n:a:h" opt; do
  case $opt in
    n)
      NS=$OPTARG
      ;;
    a)
      action=$OPTARG
      found=false
      for action_index in $actions; do
          [ "$action" == "$action_index" ] && found=true
      done
      [ "$found" == "false" ] && usage
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
echo "Working in namespace [$NS]."
nsf="--namespace ${NS}" # namespace flag for kubectl command

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl > /dev/null 2>&1 || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
. $UTILS # include utils for wait function and status
[ -z "$action" ] && usage

nsf="--namespace ${NS}" # namespace flag for kubectl command


# Main
# Execute the action function
$action


