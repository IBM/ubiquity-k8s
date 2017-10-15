#!/bin/bash -ex

# -------------------------------------------------------------------------
# This is internal script of how to install flex on minions
# TODO : Later on a the flexvolume will be deployed by daemon set
#
# Prerequisites  : ssh passwordless to minions, kubectl, and the flex file in the same directory of this script
# Args : if restart-kubelet given then the script will also restart the kubelets on the nodes
# -------------------------------------------------------------------------


scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
UTILS=$scripts/ubiquity_utils.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-database
FLEX_DIRECTORY='/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex'
flexfile_name=ubiquity-k8s-flex
flexfile_path=$scripts/${flexfile_name}
flexconf_name=${flexfile_name}.conf
ARG_VALUE_TO_RESTART_KUBELET="restart-kubelet"

# Validations
[ ! -d "$YML_DIR" ] && { echo "Error: YML directory [$YML_DIR] does not exist."; exit 1; }
which kubectl || { echo "Error: kubectl not found in PATH"; exit 2; }
[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
#### [ ! -f ${flexfile_path} ] && { echo "Error: ./${flexfile_path} not found."; exit 4; }
[ -z "$1" ] && kubelet_restart="no" || kubelet_restart="$1"


ubiquity_service_ip=`kubectl get svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
echo "The IP of ubiquity service is : $ubiquity_service_ip."

for minion in `kubectl get  nodes  | sed '1d' | awk '{print $1}'`; do
#    echo "======= ($minion) Deploying flexvolume... ======="
#    ssh root@${minion} "sudo mkdir -p ${FLEX_DIRECTORY}"
#    scp ${flexfile_path} root@${minion}:${FLEX_DIRECTORY}
#    ssh root@${minion} "sudo chmod u+x ${FLEX_DIRECTORY}/${flexfile_name}"

#    echo "====== ($minion) Create ubiquity-k8s-flex.conf ======"
#    ssh root@${minion} "mkdir -p /etc/ubiquity"
    ssh root@${minion} "cat > /etc/ubiquity/${flexconf_name} << EOF
logPath = \"/tmp\"
backends = [\"scbe\"]
logLevel = \"debug\"

[UbiquityServer]
address = \"${ubiquity_service_ip}\"
port = 9999
EOF
"
    echo ""
    if [ "$kubelet_restart" = "${ARG_VALUE_TO_RESTART_KUBELET}" ]; then
        echo "====== ($minion) restart kubelet... ======"
        ssh root@${minion} 'sudo [ -d /tmp/ubiquity ] && echo "log path: /tmp/ubiquity exist" || mkdir -p /tmp/ubiquity ; echo "create log path: /tmp/ubiquity; "'
        ssh root@${minion} 'sudo systemctl restart kubelet'
        ssh root@${minion} 'sudo systemctl restart docker' # TODO not sure if its needed

        echo "====== ($minion) disable firewall... ======"
        ssh root@${minion} 'systemctl stop firewalld' || : # TODO remove it, since the user must set the port in advance.
    else
       echo "====== ($minion) SKIP restart kubelet ======"
    fi
done


if [ "$kubelet_restart" = "${ARG_VALUE_TO_RESTART_KUBELET}" ]; then
    # TODO just make sure the kubelet is up (in one minion, later on we should make sure its ok for all of them)
    one_minion=`kubectl get nodes | sed '1d' | awk '{print $1}'| tail -1`
    echo "kubelet status on ${one_minion}"
    ssh root@${one_minion} 'service kubelet status || :'
    echo ""
    echo "sleep 30 and then check again the status of the kubelet"
    sleep 30
    ssh root@${one_minion} 'service kubelet status || :'
    ssh root@${one_minion} 'kubectl get nodes || :'
else
   echo "====== SKIP wait for kubelet, because we did not restart the kubelet ======"
fi


echo ""
echo "Finished to install flex."