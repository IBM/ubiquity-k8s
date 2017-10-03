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
# The script install all Ubiquity components inside kubernetes(k8s) cluster.
#
# Create the following components:
#   1. Ubiquity service
#   2. Ubiquity-db service
#   3. Ubiquity deployment
#   4. configmap k8s-config for Ubiquity-k8s-provisioner  (if already exist then skip creation)
#   5. Ubiquity-k8s-provisioner deployment
#   6. Storage class that match for the UbiquityDB PVC
#   7. PVC for the UbiquityDB
#   8. ubiquity-k8s-flex DaemonSet (deploy the flex + flex config file to all nodes)
#   9. optional : ubiquity-db deployment (only if -d flag specify. Use it only if the flex already deployed in the past on the nodes,
#                 therefor no restart for kubelet is needed before run the deployment)
#
# Prerequisites to run this test:
#   - The script assumes all the yamls exist under ./yamls and updated with relevant configuration.
#     If config file (with placeholder KEY=VALUE) given as argument then the script will apply all the place holders inside all the yaml files.
#   - Run the script in the Kubernetes master node where you have access to kubectl command.
#   - See usage function below for more details about flags.
# -------------------------------------------------------------------------

function usage()
{
  cmd=`basename $0`
  echo "USAGE   $cmd [-c file] [-k file] [-s create-ubiquity-db] [-d] [-h]"
  echo "  Options:"
  echo "    -c [file] : The script will update all the yamls with the place holders in the file"
  echo "    -k [file] : Kubernetes config file for ubiquity-k8s-provisioner (default is ~/.kube/config)"
  echo "    -s create-ubiquity-db :"
  echo "         Create only the ubiquity-db deployment and wait for its creation."
  echo "         Uses this option after the kubelet on the nodes were restarted."
  echo "    -d : Use it only if you already have deployed flex on the nodes, so ubiquity-db will be created as well."
  echo "    -h : Display this usage"
  exit 1
}

function update_ymls_with_playholders()
{
   placeholder_file="$1"
   [ -z "${placeholder_file}" ] && return

   [ ! -f "${placeholder_file}" ] && { echo "Error : ${placeholder_file} not found."; exit 3; }
   read -p "Updating ymls with placeholders from ${placeholder_file} file. Are you sure (y/n): " yn
   if [ "$yn" = "y" ]; then
       for line in `cat ${placeholder_file}`; do
          echo "   $line"
          placeholder=`echo $line | awk -F= '{print $1}'`
          value=`echo $line | awk -F= '{print $2}'`
          sed -i "s|${placeholder}|${value}|g" "$YML_DIR"/*.yml
          sed -i "s|${placeholder}|${value}|g" "$SANITY_YML_DIR"/*.yml
       done
       echo "Finish to update yaml according to ${placeholder_file}"
   else
      echo "Running installation without update ymls with placeholder."
   fi
}

function create_only_ubiquity_db_deployment()
{
    echo "Creating ubiquity-db deployment... (Assume flex plugin was already loaded on all the nodes)"
    kubectl create -f ${YML_DIR}/ubiquity-db-deployment.yml
    echo "Waiting for deployment [ubiquity-db] to be created..."
    wait_for_deployment ubiquity-db 40 5
    echo ""
    echo "Ubiquity installation finished successfully in the Kubernetes cluster. (To list ubiquity deployments run $> ./ubiquity_deployments.sh status)"
}

# Variables
scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
SANITY_YML_DIR="$scripts/yamls/sanity_yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
K8S_CONFIGMAP_FOR_PROVISIONER=k8s-config

# Handle flags
to_deploy_ubiquity_db="false"
KUBECONF=~/.kube/config
CONFIG_SED_FILE=""
CREATE_ONLY_UBIQUITY_DB=""
while getopts ":dc:k:s:h" opt; do
  case $opt in
    d)
      to_deploy_ubiquity_db="true"
      ;;
    c)
      CONFIG_SED_FILE=$OPTARG
      ;;
    k)
      KUBECONF=$OPTARG
      ;;
    s)
      CREATE_ONLY_UBIQUITY_DB=$OPTARG
      [ "$CREATE_ONLY_UBIQUITY_DB" != "create-ubiquity-db" ] && usage
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

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl > /dev/null 2>&1 || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
[ ! -f "$KUBECONF" ] && { echo "$KUBECONF not found"; exit 2; }
[ -n "$CONFIG_SED_FILE" -a ! -f "$CONFIG_SED_FILE" ] && { echo "$CONFIG_SED_FILE not found"; exit 2; }
. $UTILS # include utils for wait function and status

if [ -n "$CREATE_ONLY_UBIQUITY_DB" ]; then
    create_only_ubiquity_db_deployment
    exit 0
fi

# update yamls if -c flag given
[ -n "${CONFIG_SED_FILE}" ] && update_ymls_with_playholders ${CONFIG_SED_FILE}

# Start to create ubiquity components in order
# --------------------------------------------
# TODO need to create namespace

kubectl create -f ${YML_DIR}/ubiquity-service.yml
kubectl create -f ${YML_DIR}/ubiquity-db-service.yml
kubectl create -f ${YML_DIR}/ubiquity-deployment.yml
wait_for_deployment ubiquity 5 5

if ! kubectl get cm/${K8S_CONFIGMAP_FOR_PROVISIONER} > /dev/null 2>&1; then
    kubectl create configmap ${K8S_CONFIGMAP_FOR_PROVISIONER} --from-file $KUBECONF
else
    echo "Skip the creation of ${K8S_CONFIGMAP_FOR_PROVISIONER} configmap, because its already exist"
fi
kubectl create -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
wait_for_deployment ubiquity-k8s-provisioner 5 5

# Create Storage class and PVC, then wait for PVC and PV creation
kubectl create -f ${YML_DIR}/storage-class.yml
kubectl create -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for ${UBIQUITY_DB_PVC_NAME} PVC to be created"
wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 60 5
pvname=`kubectl get pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
echo "Waiting for ${pvname} PV to be created"
wait_for_item pv $pvname ${PVC_GOOD_STATUS} 20 3

# Create ${flex_conf} configmap with the ubiquity-service clusterIP
ubiquity_service_ip=`kubectl get svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
echo "Deploy flex driver as infinit daemonset, Its also copy the flex config file with the ubiquity service IP [$ubiquity_service_ip]"
flex_conf="ubiquity-k8s-flex.conf"
sed -i "s/address = .*/address = \"${ubiquity_service_ip}\"/"  ${YML_DIR}/${flex_conf}
kubectl create configmap ${flex_conf} --from-file ${YML_DIR}/${flex_conf}

kubectl create -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml

if [ "${to_deploy_ubiquity_db}" == "true" ]; then
    create_only_ubiquity_db_deployment
else
    echo ""
    echo "Ubiquity installation finished, but Ubiquity is NOT ready yet."
    echo "  You must do : (1) Manually restart kubelet service on all minions to reload the new flex driver"
    echo "                (2) Deploy ubiquity-db by      $> $0 -s create-ubiquity-db"
    echo "                Note : View ubiquity status by $> ./ubiquity_deployments.sh status"
    echo ""
fi
