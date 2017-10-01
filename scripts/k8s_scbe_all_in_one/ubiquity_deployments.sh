#!/bin/bash -e

# -------------------------------------------------------------------------
# The script start\stop ubiquity deployments with the right order.
# The right order is to start ubiquity and provisioner before ubiquity-db and vise versa for stop.
# The script assume that you already installed ubiquity via the ubiquity_install.sh script.
#
# Start flow order:
#   - Validate ubiquity pvc exist (if not exit)
#   - Create Ubiquity deployment
#   - Create Ubiquity-k8s-provisioner deployment
#   - Create ubiquity-db-deployment deployment
#
# Start flow order:
#   - Delete ubiquity-db-deployment deployment
#   - Delete Ubiquity-k8s-provisioner deployment
#   - Delete Ubiquity deployment
#
# Assuming the following components was already created during the ubiquity_install.sh script.
#   - Ubiquity service
#   - Ubiquity-db service
#   - configmap k8s-config for Ubiquity-k8s-provisioner
#   - Storage class that match for the UbiquityDB PVC
#   - PVC for the UbiquityDB
#   - FlexVolume already installed on each minion and configured well
#
# -------------------------------------------------------------------------


function usage()
{
   echo "Usage, $0 [start|stop|status|statusall|help]"
   echo "    Options"
   echo "       start  : Create ubiquity, provisioner deployments, flex daemonset and ubiquity-db deployments"
   echo "       stop   : Delete ubiquity-db(wait for deletion), provisioner deployment, flex daemonset, ubiquity deployment"
   echo "       status : kubectl get to all the ubiquity components"
   echo "       statusall : kubectl get configmap,storageclass,pv,pvc,service,daemonset,deployment,pod"
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
    sleep 5 # TODO wait for deployment
    kubectl create -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl create -f ${YML_DIR}/ubiquity-db-deployment.yml
    echo "Finished to start ubiquity components. Run $0 status to get more details."
}

function stop()
{
    # TODO delete the deployments only if its actually exist
    kubectl delete -f $YML_DIR/ubiquity-db-deployment.yml
    sleep 30 # TODO wait till deployment stopped including the POD.
    kubectl delete -f $YML_DIR/ubiquity-k8s-provisioner-deployment.yml
    kubectl delete -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    kubectl delete -f $YML_DIR/ubiquity-deployment.yml
    echo "Finished to stop ubiquity deployments. Run $0 status to get more details."
}


function status()
{
    # kubectl get on configmap, storageclass, deployment and pod that related to ubiquity
    kubectl get configmap k8s-config || :
    echo ""
    kubectl get storageclass | egrep "ubiquity|^NAME"  || :
    echo ""
    kubectl get pv/ibm-ubiquity-db pvc/ibm-ubiquity-db svc/ubiquity svc/ubiquity-db  daemonset/ubiquity-flex deploy/ubiquity deploy/ubiquity-db deploy/ubiquity-k8s-provisioner  || :
    echo ""
    kubectl get pod | egrep "^ubiquity|^NAME" || :
}

function statusall()
{
    kubectl get configmap,storageclass,pv,pvc,service,daemonset,deployment,pod || :
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
[ $action != "start" -a $action != "stop" -a $action != "status" -a $action != "statusall" -a $action != "help" ] && usage



$1


