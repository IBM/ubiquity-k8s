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
  echo "USAGE   $cmd [-c <file>] [-k <file>] [-s [create-ubiquity-db|create-services]] [-n <namespace>] [-d] [-h]"
  echo "  Options:"
  echo "    -c [<file>] : The script will update all the yamls with the place holders in the file"
  echo "    -k [<file>] : Kubernetes config file for ubiquity-k8s-provisioner (default is ~/.kube/config)"
  echo "    -s [STEP]"
  echo "       create-ubiquity-db :"
  echo "         Creates only the ubiquity-db deployment and waits for its creation."
  echo "         Use this option after the kubelets on the nodes have restarted."
  echo "       create-services :"
  echo "         Creates ubiquity namespace and ubiqutiy\ubiquity-db k8s service only."
  echo "         Use this option only if you want to get the IPs of ubiquity and ubiqutiy-db in advance before installation."
  echo "    -n [<namespace>] : Namespace to install ubiquity. By default its \"ubiquity\" namespace"
  echo "    -d : Use it only if you already have deployed flex on the nodes, so ubiquity-db will be created as well."
  echo "    -t [<certificates-directory>] : Create only the Ubiquity secrets\configmap for dedicated certificates."
  echo "         Create secrets ubiquity-private-certificate and ubiquity-db-private-certificate"
  echo "         Create configmap ubiquity-public-certificates"
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
    _ns=$1
    echo "Creating ubiquity-db deployment... (Assume flex plugin was already loaded on all the nodes)"
    kubectl create --namespace $_ns -f ${YML_DIR}/ubiquity-db-deployment.yml
    echo "Waiting for deployment [ubiquity-db] to be created..."
    wait_for_deployment ubiquity-db 40 5 $_ns
    echo ""
    echo "Ubiquity installation finished successfully in the Kubernetes cluster. (To list ubiquity deployments run $> ./ubiquity_deployments.sh -a status -n $NS)"
}

function create_only_namespace_and_services()
{
    # This function create ubiquity name space, ubiquity service and ubiquity-db service (if not exist)

    # Create ubiquity namespace
    if [ "$NS" = "$UBIQUITY_DEFAULT_NAMESPACE" ]; then
        if ! kubectl get namespace $UBIQUITY_DEFAULT_NAMESPACE >/dev/null 2>&1; then
           kubectl create -f ${YML_DIR}/ubiquity-namespace.yml
        else
           echo "$UBIQUITY_DEFAULT_NAMESPACE already exist. (Skip namespace creation)"
        fi
    fi

    # Create ubiquity service
    if ! kubectl get $nsf service ${UBIQUITY_SERVICE_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-service.yml
    else
       echo "$UBIQUITY_SERVICE_NAME service already exist. (Skip service creation)"
    fi

    # Create ubiquity-db service
    if ! kubectl get $nsf service ${UBIQUITY_DB_SERVICE_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-db-service.yml
    else
       echo "$UBIQUITY_DB_SERVICE_NAME service already exist. (Skip service creation)"
    fi
}

function create_private_certificates_secrets_and_public_as_configmap()
{
    echo "Creating secrets [ubiquity-private-certificate and ubiquity-db-private-certificate] and configmap [ubiquity-public-certificates] based from directory $certDir"
    certDir="$1"  # directory of the certificates
    [ ! -d "$certDir" ] && { echo "Error: $certDir directory not found."; exit 2; }

    # Validation all cert files in the certDir directory
    expected_cert_files="ubiquity.key ubiquity.crt ubiquity-db.key ubiquity-db.crt ubiquity-trusted-ca.crt ubiquity-db-trusted-ca.crt scbe-trusted-ca.crt "
    for certfile in $expected_cert_files; do
        [ ! -f $certDir/$certfile ] && { echo "Missing cert file $certDir/$certfile"; echo "Mandatory certificate files are : $expected_cert_files"; exit 2; }
    done

    # Validation secrets and configmap not exist before creation
    kubectl get secret $nsf ubiquity-db-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-db-private-certificate]" || :
    kubectl get secret $nsf ubiquity-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-private-certificate]" || :
    kubectl get configmap $nsf ubiquity-public-certificates >/dev/null 2>&1 && already_exist "configmap [ubiquity-public-certificates]" || :

    cd $certDir
    kubectl create secret $nsf generic ubiquity-db-private-certificate --from-file=ubiquity-db.key --from-file=ubiquity-db.crt
    kubectl create secret $nsf generic ubiquity-private-certificate --from-file=ubiquity.key --from-file=ubiquity.crt
    kubectl create configmap $nsf ubiquity-public-certificates --from-file=ubiquity-db-trusted-ca.crt=ubiquity-db-trusted-ca.crt --from-file=scbe-trusted-ca.crt=scbe-trusted-ca.crt --from-file=ubiquity-trusted-ca.crt=ubiquity-trusted-ca.crt
    cd -

    kubectl get $nsf secrets/ubiquity-db-private-certificate secrets/ubiquity-private-certificate cm/ubiquity-public-certificates
    echo ""
    echo "Finished to create secrets and configmap for Ubiquity certificates."
}
function already_exist() { echo "Error: Secret $1 is already exist. Please delete it first."; exit 2; }

# Variables
scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
SANITY_YML_DIR="$scripts/yamls/sanity_yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
K8S_CONFIGMAP_FOR_PROVISIONER=k8s-config

[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
. $UTILS # include utils for wait function and status

NS="$UBIQUITY_DEFAULT_NAMESPACE" # Set as the default namespace

# Handle flags
to_deploy_ubiquity_db="false"
KUBECONF=~/.kube/config
CONFIG_SED_FILE=""
STEP=""
CERT_DIR=""
while getopts ":dc:k:s:n:t:h" opt; do
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
      STEP=$OPTARG
      [ "$STEP" != "create-ubiquity-db" -a "$STEP" != "create-services" ] && usage
      ;;
    n)
      NS=$OPTARG
      ;;
    t)
      CERT_DIR=$OPTARG
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

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl > /dev/null 2>&1 || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f "$KUBECONF" ] && { echo "$KUBECONF not found"; exit 2; }
[ -n "$CONFIG_SED_FILE" -a ! -f "$CONFIG_SED_FILE" ] && { echo "$CONFIG_SED_FILE not found"; exit 2; }

if [ -n "$CERT_DIR" ]; then
   create_private_certificates_secrets_and_public_as_configmap "$CERT_DIR"
   exit 0
fi

echo "Install Ubiquity in namespace [$NS]..."

if [ "$STEP" = "create-ubiquity-db" ]; then
    create_only_ubiquity_db_deployment $NS
    exit 0
fi

if [ "$STEP" = "create-services" ]; then
    echo "Partially install Ubiquity - creates only the ubiquity and ubiquity-db services up."
    create_only_namespace_and_services $NS

    read -p "Going to update Ubiquity yamls to support certificates via volumes(secets and configmap). Are you sure (y/n): " yn
    if [ "$yn" = "y" ]; then
       # this sed just removes the comments from all the certificates lines on the ymls
       ymls_to_updates="${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml ${YML_DIR}/ubiquity-deployment.yml ${YML_DIR}/ubiquity-db-deployment.yml"
       sed -i 's/^# Cert #\(.*\)/\1  # Cert #/g' ${ymls_to_updates}
       echo "yamls were updated to support certificates ($ymls_to_updates)"
    else
       echo "yamls were NOT updated to support certificates."
    fi

    kubectl get $nsf svc/ubiquity svc/ubiquity-db
    echo ""
    echo "Finish to create namespace, ${UBIQUITY_SERVICE_NAME} service and ${UBIQUITY_DB_SERVICE_NAME} service"
    echo "Attention: To complete Ubiqutiy installation do :"
    echo "   Prerequisite"
    echo "     (1) Generate dedicated certificates for ubiquity, ubiquity-db and scbe. (with specific name files)"
    echo "     (2) Create secrets and configmap to store the certificates and trusted CA files, by just running :"
    echo "          $> $0 -t <certificates-directory> -n $NS"
    echo "           It will create the following:"
    echo "           secret1 named [ubiquity-private-certificate]"
    echo "           secret2 named [ubiquity-db-private-certificate]"
    echo "           configmap named [ubiquity-public-certificates]"
    echo "   Complete the installation:"
    echo "     (1)  $> $0 -c <file> -n $NS"
    echo "     (2)  Manually restart kubelet service on all minions to reload the new flex driver"
    echo "     (3)  $> $0 -s create-ubiquity-db -n $NS"
    echo ""
    exit 0
fi

# update yamls if -c flag given
[ -n "${CONFIG_SED_FILE}" ] && update_ymls_with_playholders ${CONFIG_SED_FILE}



# Start to create ubiquity components in order
# --------------------------------------------
create_only_namespace_and_services

kubectl create $nsf -f ${YML_DIR}/ubiquity-deployment.yml
wait_for_deployment ubiquity 20 5 $NS

if ! kubectl get $nsf cm/${K8S_CONFIGMAP_FOR_PROVISIONER} > /dev/null 2>&1; then
    kubectl create $nsf configmap ${K8S_CONFIGMAP_FOR_PROVISIONER} --from-file $KUBECONF
else
    echo "Skip the creation of ${K8S_CONFIGMAP_FOR_PROVISIONER} configmap, because its already exist"
fi
kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
wait_for_deployment ubiquity-k8s-provisioner 20 5 $NS

# Create Storage class and PVC, then wait for PVC and PV creation
kubectl create $nsf -f ${YML_DIR}/storage-class.yml
kubectl create $nsf -f ${YML_DIR}/ubiquity-db-pvc.yml
echo "Waiting for ${UBIQUITY_DB_PVC_NAME} PVC to be created"
wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 60 5 $NS
pvname=`kubectl get $nsf pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
echo "Waiting for ${pvname} PV to be created"
wait_for_item pv $pvname ${PVC_GOOD_STATUS} 20 3 $NS

# Create ${flex_conf} configmap with the ubiquity-service clusterIP
ubiquity_service_ip=`kubectl get $nsf svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
echo "Deploy flex driver as infinit daemonset, Its also copy the flex config file with the ubiquity service IP [$ubiquity_service_ip]"
flex_conf="ubiquity-k8s-flex.conf"
sed -i "s/address = .*/address = \"${ubiquity_service_ip}\"/"  ${YML_DIR}/${flex_conf}
kubectl create $nsf configmap ${flex_conf} --from-file ${YML_DIR}/${flex_conf}

kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml

if [ "${to_deploy_ubiquity_db}" == "true" ]; then
    create_only_ubiquity_db_deployment $NS
else
    echo ""
    echo "Ubiquity installation finished, but Ubiquity is NOT ready yet."
    echo "  You must do : (1) Manually restart kubelet service on all minions to reload the new flex driver"
    echo "                (2) Deploy ubiquity-db by      $> $0 -s create-ubiquity-db -n $NS"
    echo "                Note : View ubiquity status by $> ./ubiquity_deployments.sh -a status -n $NS"
    echo ""
fi
