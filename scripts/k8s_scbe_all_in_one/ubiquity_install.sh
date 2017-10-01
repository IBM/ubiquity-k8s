#!/bin/bash -e

# -------------------------------------------------------------------------
# The script install all Ubiquity components inside kubernetes(k8s) cluster.
#
# Create the following components:
#   - Ubiquity service
#   - Ubiquity-db service
#   - Ubiquity deployment
#   - configmap k8s-config for Ubiquity-k8s-provisioner
#   - Ubiquity-k8s-provisioner deployment
#   - Storage class that match for the UbiquityDB PVC
#   - PVC for the UbiquityDB
#   - Update all minions with the Ubiquity service IP # TODO will be change to daemon set later on
#   - TODO : create postgres deployment
#
# Prerequisites to run this test:
#   - The script assumes all the yamls exist under ./yamls and updated with relevant configuration.
#     If config file (with placeholder KEY=VALUE) given as argument then the script will apply all the place holders inside all the yaml files.
#   - The script assumes the customer update all the ymls with relevant configurations.
#   - Run the script in the Kubernetes master node where you have access to kubectl command.
#   - KUBECONF environment should set to the k8s config file that will be loaded as configmap for the ubiquity provisioner.
# -------------------------------------------------------------------------

function update_ymls_with_playholders()
{
   placeholder_file="$1"
   [ -z "${placeholder_file}" ] && return

   [ ! -f "${placeholder_file}" ] && { echo "Error : ${placeholder_file} not found."; exit 3; }
   read -p "Update ymls with placeholders. Are you sure (y/n): " yn
   if [ "$yn" = "y" ]; then
       for line in `cat ${placeholder_file}`; do
          echo "   $line"
          placeholder=`echo $line | awk -F= '{print $1}'`
          value=`echo $line | awk -F= '{print $2}'`
          sed -i "s|${placeholder}|${value}|g" "$YML_DIR"/*.yml
       done
       echo "Finish to update yaml according to ${config_keys}"
   else
      echo "Running installation without update ymls with placeholder."
   fi
}

function usage()
{
  cmd=`basename $0`
  echo "USAGE, $cmd [-d] [-c filepath] [-k filepath] [-h]"
  echo "  Options:"
  echo "    -d            : Assume flex plugin in already loaded so create ubiquity-db deployment after daemon-set"
  echo "    -c [filepath] : if given, then first sed all key-value in filepath on all yaml directory"
  echo "    -k [filepath] : Kubernetes config file for ubiquity-k8s-provisioner (default is ~/.kube/config)"
  echo "    -h            : Display this usage"
  exit 1
}

scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
FLEX_DIRECTORY='/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex'

# Handle flags to the installer
to_deploy_ubiquity_db="false"
KUBECONF=~/.kube/config
CONFIG_SED_FILE=""
while getopts ":dc:k:h" opt; do
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
which kubectl || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
[ ! -f "$KUBECONF" ] && { echo "$KUBECONF not found"; exit 2; }
[ -n "$CONFIG_SED_FILE" -a ! -f "$CONFIG_SED_FILE" ] && { echo "$CONFIG_SED_FILE not found"; exit 2; }
. $UTILS # include utils for wait function and status


[ -n "${CONFIG_SED_FILE}" ] && update_ymls_with_playholders ${CONFIG_SED_FILE}

# Start to create ubiquity components in order
# --------------------------------------------
# TODO need to create namespace

kubectl create -f ${YML_DIR}/ubiquity-service.yml
kubectl create -f ${YML_DIR}/ubiquity-db-service.yml

kubectl create -f ${YML_DIR}/ubiquity-deployment.yml
sleep 2 # TODO wait for deployment

kubectl create configmap k8s-config --from-file $KUBECONF   # TODO consider use secret or assume customer set it up in advance
sleep 2 # TODO wait for deployment

kubectl create -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml

kubectl create -f ${YML_DIR}/storage-class.yml

kubectl create -f ${YML_DIR}/ubiquity-db-pvc.yml

echo "Waiting for ${UBIQUITY_DB_PVC_NAME} PVC to be created"
wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 60 5
pvname=`kubectl get pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
echo "Waiting for ${pvname} PV to be created"
wait_for_item pv $pvname ${PVC_GOOD_STATUS} 20 3

ubiquity_service_ip=`kubectl get svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
echo "Deploy flex driver as infinit daemonset, Its also copy the flex config file with the ubiquity service IP [$ubiquity_service_ip]"
flex_conf="ubiquity-k8s-flex.conf"
sed -i "s/address = .*/address = \"${ubiquity_service_ip}\"/"  ${YML_DIR}/${flex_conf}
kubectl create configmap ${flex_conf} --from-file ${YML_DIR}/${flex_conf}

kubectl create -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml

if [ "${to_deploy_ubiquity_db}" == "true" ]; then
    echo "Assume flex plugin in already loaded in minions, lets create the ubiquity-db deployment..."
    kubectl create -f ${YML_DIR}/ubiquity-db-deployment.yml
    wait_for_deployment ubiquity-db 40 5

    echo ""
    echo "Finished to install Ubiquity components inside the Kubernetes cluster."
    echo "  To list ubiquity deployments :$> ubiquity_deployments.sh status"
else
    echo ""
    echo "Finished, but ubiquity not ready yet."
    echo " You must do : (1) Manually restart kubelet service on all minions to reload the new flex driver"
    echo "               (2) $> kubectl create -f ubiquity-db-deployment.yml"
    echo ""
fi