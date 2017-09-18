#!/bin/bash -ex

# -------------------------------------------------------------------------
# The script install all Ubiquity components inside kubernetes(k8s) cluster.
#
# Create the following components:
#   - Ubiquity service
#   - UbiquityDB service
#   - Ubiquity deployment
#   - configmap k8s-config for Ubiquity-k8s-provisioner
#   - Ubiquity-k8s-provisioner deployment
#   - TODO : Create Storage class that will match for the UbiquityDB PVC
#   - TODO : Create the PVC for the UbiquityDB
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

YML_DIR="./yamls"

[ -z "$KUBECONF" ] && KUBECONF=~/.kube/config || :
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl || { echo "Error: kubectl not found in PATH"; exit 2; }

update_ymls_with_playholders "$1"

# TODO need to create namespace

kubectl create -f ${YML_DIR}/ubiquity-service.yml
kubectl create -f ${YML_DIR}/ubiquity-db-service.yml

kubectl create -f ${YML_DIR}/ubiquity-deployment.yml
sleep 2 # TODO wait for deployment

kubectl create configmap k8s-config --from-file $KUBECONF   # TODO consider use secret or assume customer set it up in advance
sleep 2 # TODO wait for deployment

kubectl create -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml

# TODO : Create Storage class that will match for the Ubiquity-db PVC
# TODO : Create the PVC for the Ubiquity-db

ubiquity_service_ip=`kubectl get svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`

# update the flex with the new IP of the ubiquity service  # TODO will be done in daemon set later
flex_conf=/etc/ubiquity/ubiquity-k8s-flex.conf
for minion in `kubectl get  nodes  | sed '1d' | awk '{print $1}'`; do
   echo "update $minion flex with IP=${ubiquity_service_ip}"
   ssh root@$minion "sed -i 's/^address =.*/address = \"${ubiquity_service_ip}\"/' $flex_conf"
   ssh root@$minion "grep address $flex_conf"
done


# TODO create postgres deployment

echo ""
echo "Ubiquity is ready to use."