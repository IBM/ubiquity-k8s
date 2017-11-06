#!/bin/bash -e

# -------------------------------------------------------------------------
# "IBM Storage Enabler for Containers" cli tool.
# The tool is an helper to do the following actions : start, stop, status, collect_logs, sanity.
# The script assume that you already installed ubiquity via the ubiquity_install.sh script.
#
# "start" action flow:
#   1. Validate ubiquity pvc exist (exit if not)
#   2. Create ubiquity deployment
#   3. Create ubiquity-k8s-provisioner deployment
#   4. Create ubiquity-k8s-flex Daemonset
#   5. Create ubiquity-db-deployment deployment  (This step uses PVC, so ubiqiuty must be up and ready for it)
#
# "stop" action flow:
#   1. Delete ubiquity-db-deployment deployment
#   2. Delete Ubiquity-k8s-provisioner deployment
#   3. Create ubiquity-k8s-flex Daemonset
#   4. Delete Ubiquity deployment
#
# See Usage for more detail.
# -------------------------------------------------------------------------


function usage()
{
   echo "Usage, $0 -a [`echo $actions | sed 's/ /|/g'`] [-n <namespace>] [-h]"
   echo " -a <action>"
   echo "   start      : Create ubiquity, provisioner deployments, flex daemonset and ubiquity-db deployments"
   echo "   stop       : Delete ubiquity-db (wait for deletion), provisioner, flex daemonset and ubiquity deployment"
   echo "   status     : Display all ubiquity components"
   echo "   statuswide : Display all ubiquity components (-o wide flag)"
   echo "   getall     : Display related components on <namespace> and the default namespaces."
   echo "   getallwide : getall with more details (-o wide)"
   echo "   collect_logs : Create a directory with all Ubiquity logs"
   echo "   sanity       : This is a sanity test - create and delete pvc and pod."
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
    echo "Make sure ${UBIQUITY_DB_PVC_NAME} exists and bound to PV (exit otherwise)..."
    wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 5 2 $NS

    kubectl create $nsf -f ${YML_DIR}/ubiquity-deployment.yml
    wait_for_deployment ubiquity 20 5 $NS

    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
    wait_for_deployment ubiquity-k8s-provisioner 20 5 $NS

    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl create $nsf -f ${YML_DIR}/ubiquity-db-deployment.yml
    wait_for_deployment ubiquity-db 40 5 $NS
    echo "Finished to start ubiquity components. Note : View ubiquity status by : $> $0 -a status -n $NS"
}

function stop()
{
    kubectl_delete="kubectl delete $nsf --ignore-not-found=true"

    # TODO Instead of using yml file to delete object, we can just delete them by object name
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
    echo "Finish collecting \"$PRODUCT_NAME\" logs inside directory -> $logdir"
}


function status()
{
    # kubectl get on all the ubiquity components, if one of the components are not found
    rc=0
    flags="$1"

    cmd="kubectl get $flags storageclass | egrep \"ubiquity|^NAME\""
    echo $cmd
    echo '---------------------------------------------------------------------'
    kubectl get $flags  storageclass | egrep "ubiquity|^NAME" || rc=$?
    echo ""

    cmd="kubectl get $nsf $flags secret/ubiquity-db-credentials secret/scbe-credentials cm/k8s-config cm/ubiquity-k8s-flex.conf cm/ubiquity-configmap pv/ibm-ubiquity-db pvc/ibm-ubiquity-db svc/ubiquity svc/ubiquity-db  daemonset/ubiquity-k8s-flex deploy/ubiquity deploy/ubiquity-db deploy/ubiquity-k8s-provisioner"
    echo $cmd
    echo '---------------------------------------------------------------------'
    $cmd  || rc=$?

    echo ""
    cmd="kubectl get $nsf $flags  pod | egrep \"^ubiquity|^NAME\""
    echo $cmd
    echo '---------------------------------------------------------------------'
    kubectl get $nsf $flags  pod | egrep "^ubiquity|^NAME" || rc=$?

    if [ $rc != 0 ]; then
       echo ""
       echo "Ubiquity status [NOT ok]. Some components are missing(review the output above)"
       exit 5
    else
      kubectl get $nsf pod | egrep "^ubiquity" | grep -v Running > /dev/null 2>&1 && rc=$? || rc=$?
      if [ $rc = 0 ]; then
          echo ""
          echo "Ubiquity status [NOT ok]. Some Pods are NOT in Running state (review the output above)"
          exit 5
      fi
    fi
}

function statuswide()
{
    status "-o wide"
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
    echo "\"$PRODUCT_NAME\" sanity finished successfully."
}


scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
SANITY_YML_DIR="$scripts/yamls/sanity_yamls"
UTILS=$scripts/ubiquity_lib.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
FLEX_DIRECTORY='/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex'
actions="start stop status statuswide getall getallwide collect_logs sanity"

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


