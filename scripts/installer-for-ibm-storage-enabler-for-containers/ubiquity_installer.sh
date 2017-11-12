#!/bin/bash -e

# -------------------------------------------------------------------------
# "IBM Storage Enabler for Containers" installer script that installs
# the following components inside the kubernetes cluster:
#       "IBM Storage Enabler for Containers"
#       "IBM Storage Dynamic Provisioner for Kubernetes"
#       "IBM Storage Flex Volume for Kubernetes"
#
# How to Run the installer:
# ==========================================
# Preparations:
#  1. MANUEL operation : update the ubiquity.config with the relevant VALUEs.
#  2. Only for SSL_MODE=verify-full, you first need to run the following steps:
#       2.1. If you need the IP of ubiquity to generate certificates then run this command:
#           $> ./ubiquity_installer.sh -s create-services
#       2.2. MANUEL operation: creates a dedicated certificates for ubiquity, ubiquity-db and SCBE. (not responsibility of this script)
#       2.3. Generate kubernetes secrets that will holds the certificates created in #2.2 step, by running the command:
#           $> ./ubiquity_installer.sh -s create-secrets-for-certificates -t <certificates-directory>
#
#  3. Update the ymls with the key=values from the ubiquity.config file, by running the command:
#    $> ./ubiquity_installer.sh -s update-ymls -c ubiquity.config
#
# Installation:
#  1. Install the solution (without ubiquity-db):
#    $> ./ubiquity_installer.sh -s install -k <k8s-config-file>
#    NOTE : <k8s-config-file> is needed for the ubiquity-k8s-provisioner to access the Kubernetes API server.
#           k8s config file usually can be found ~/.kube/config or in /etc/kubernetes directory.
#
#  2. MANUEL operation : The user MUST manual restart the kubelet on all the minions
#
#  3. Install the ubiquity-db
#     $> ./ubiquity_installer.sh -s create-ubiquity-db
#
#
# Describe the script STEPs in great detail:
# ==========================================
# STEP "update-ymls"
#   Just update all the ymls with the placeholders given in the -c <file>.
#
# STEP "install" creates the following components:
#   1. Namespace    "ubiquity"                (skip if already exist)
#   2. Service(clusterIP type) "ubiquity"     (skip if already exist)
#   3. Service(clusterIP type) "ubiquity-db"  (skip if already exist)
#   4. ConfigMap    "ubiquity-configmap"      (skip if already exist)
#   5. Secret       "scbe-credentials"        (skip if already exist)
#   6. Secret       "ubiquity-db-credentials" (skip if already exist)
#   3. Deployment   "ubiquity"
#   4. ConfigMap    "k8s-config"              (skip if already exist)
#   5. Deployment   "ubiquity-k8s-provisioner"
#   6. StorageClass <name given by user>      (skip if already exist)
#   7. PVC          "ibm-ubiquity-db"
#   8. DaemonSet    "ubiquity-k8s-flex"
#
# STEP "create-ubiquity-db" creates the following components:
#   9. Deployment   "ubiquity-db"
#
# STEP "create-services" creates the following components:
#   1. Namespace    "ubiquity"                (skip if already exist)
#   2. Service(clusterIP type) "ubiquity"     (skip if already exist)
#   3. Service(clusterIP type) "ubiquity-db"  (skip if already exist)
#
# STEP "create-secrets-for-certificates" creates the following components:
#   1. secret    "ubiquity-db-private-certificate"
#   2. secret    "ubiquity-private-certificate"
#   3. configmap "ubiquity-public-certificates"
#
# Prerequisites to run this test:
# ===============================
#   - The script assumes all the yamls exist under ./yamls and updated with relevant configuration.
#   - kubectl, base64 command lines must exist on the node you run the script.
#   - See usage function below for more details about flags.
#
# -------------------------------------------------------------------------

function usage()
{
  cmd=`basename $0`
  cat << EOF
USAGE   $cmd -s <STEP> <FLAGS>
  -s <STEP>:
    -s update-ymls -c <file>
        Replace the placeholders from -c <file> on the relevant ymls.
        Flag -c <ubiquity-config-file> is mandatory for this step
    -s install -k <k8s config file> [-n <namespace>]
        Install all ubiquity component by order (except for ubiquity-db)
        Flag -k <k8s-config-file-path> for ubiquity-k8s-provisioner.
        This file usually can be found in ~/.kube/config or in /etc/kubernetes directory.
        Flag -n <namespace>. By default its \"ubiquity\" namespace.
    -s create-ubiquity-db [-n <namespace>]
        Creates only the ubiquity-db deployment and waits for its creation.
        Use this option after install-ubiquity finished and after manual restart of the kubelets done on the nodes.

    Steps only for SSL_MODE=verify-full:
    -s create-services [-n <namespace>]
        This step creates ubiquity namespace, ubiqutiy and ubiquity-db k8s service only.
        If in order to create certificates you need to see the ubiquity and ubiqity-db IPs, then run this step.
        Flag -n <namespace>. By default its \"ubiquity\" namespace.
    -s create-secrets-for-certificates -t <certificates-directory>  [-n <namespace>]
        Creates secrets and configmap for ubiquity certificates:
            Secrets ubiquity-private-certificate and ubiquity-db-private-certificate.
            Configmap ubiquity-public-certificates.
        Flag -t <certificates-directory> that contains all the expected certificate files.
 -h : Display this usage
EOF
  exit 1
}


# STEP function
function install()
{
    ########################################################################
    ##  Install all ubiquity components in the right order and wait for creation to complete.
    ##  Fail if the deployments are already exist.
    ########################################################################

    [ -z "$KUBECONF" ] && { echo "Error: Missing -k <file> flag for STEP [$STEP]"; exit 4; } || :
    [ ! -f "$KUBECONF" ] && { echo "Error : $KUBECONF not found."; exit 3; } || :

    echo "Start to install \"$PRODUCT_NAME\"..."
    echo "Installation will be done inside the namespace [$NS]."

    create_only_namespace_and_services
    create_configmap_and_credentials_secrets

    kubectl create $nsf -f ${YML_DIR}/ubiquity-deployment.yml
    wait_for_deployment ubiquity 20 5 $NS

    echo "Creating ${K8S_CONFIGMAP_FOR_PROVISIONER} for ubiquity-k8s-provisioner from file [$KUBECONF]."
    if ! kubectl get $nsf cm/${K8S_CONFIGMAP_FOR_PROVISIONER} > /dev/null 2>&1; then
        kubectl create $nsf configmap ${K8S_CONFIGMAP_FOR_PROVISIONER} --from-file $KUBECONF
    else
        echo "Skip the creation of ${K8S_CONFIGMAP_FOR_PROVISIONER} configmap, because its already exist"
    fi
    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml
    wait_for_deployment ubiquity-k8s-provisioner 20 5 $NS

    # Create Storage class and PVC, then wait for PVC and PV creation
    if ! kubectl get $nsf -f ${YML_DIR}/storage-class.yml > /dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/storage-class.yml
    else
        echo "Skip the creation of ${YML_DIR}/storage-class.yml Storage Class, because its already exist"
    fi
    kubectl create $nsf -f ${YML_DIR}/ubiquity-db-pvc.yml
    echo "Waiting for ${UBIQUITY_DB_PVC_NAME} PVC to be created"
    wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 60 5 $NS
    pvname=`kubectl get $nsf pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
    echo "Waiting for ${pvname} PV to be created"
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 20 3 $NS

    echo "Deploy flex driver as a daemonset on all nodes and all masters.  (The daemonset will use the ubiquity service IP)"
    kubectl create $nsf -f ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml
    wait_for_daemonset ubiquity-k8s-flex 20 3 $NS

    daemonset_desiredNumberScheduled="$(get_daemonset_desiredNumberScheduled ubiquity-k8s-flex $NS)"
    number_of_nodes=`kubectl get nodes| awk '$2 ~/Ready/' | wc -l`
    flex_missing=false
    if [ "$daemonset_desiredNumberScheduled" != "$number_of_nodes" ]; then
        echo ""
        echo "*WARNING*: "
        echo "   ubiquity-k8s-flex daemonset pod MUST run on each node and master in the cluster."
        echo "   But it run only on $daemonset_desiredNumberScheduled from $number_of_nodes nodes(and masters in the cluster)."
        flex_missing=true
    fi



    if [ "${to_deploy_ubiquity_db}" == "true" ]; then
        create-ubiquity-db
    else
        echo ""
        echo "\"$PRODUCT_NAME\" installation finished, but its NOT ready yet."
        echo "  You must do : "
        [ "$flex_missing" = "true" ] && echo "     (0) ubiquity-k8s-flex daemonset pod MUST run on all nodes including all masters (Check why it does NOT)."
        echo "     (1) Manually restart kubelet service on all minions to reload the new flex driver"
        echo "     (2) Deploy ubiquity-db by      $> $0 -s create-ubiquity-db -n $NS"
        echo "     Note : View ubiquity status by $> ./ubiquity_cli.sh -a status -n $NS"
        echo ""
    fi
}

# STEP function
function update-ymls()
{
   ########################################################################
   ##  Replace all the placeholders in the config file in all the relevant yml files.
   ##  If nothing to replace then exit with error.
   ########################################################################

   # Step validation
   [ -z "${CONFIG_SED_FILE}" ] && { echo "Error: Missing -c <file> flag for STEP [$STEP]"; exit 4; } || :
   [ ! -f "${CONFIG_SED_FILE}" ] && { echo "Error : ${CONFIG_SED_FILE} not found."; exit 3; }
   which base64 > /dev/null 2>&1 || { echo "Error: base64 command not found in PATH. So cannot update ymls with base64 secret."; exit 2; }
   which egrep > /dev/null 2>&1 || { echo "Error: egrep command not found in PATH. So cannot update ymls with base64 secret."; exit 2; }

   # Validate key=value file format and no missing VALUE fill up
   egrep -v "^\s*#|^\s*$" ${CONFIG_SED_FILE} |  grep -v "^.\+=.\+$" && { echo "Error: ${CONFIG_SED_FILE} format must have only key=value lines. The above lines are with bad format."; exit 2; } || :
   grep "=VALUE$" ${CONFIG_SED_FILE} && { echo "Error: You must fill up the VALUE in ${CONFIG_SED_FILE} file."; exit 2; } || :

    # Prepare map of keys inside ${CONFIG_SED_FILE} and there associated yml file.
    UBIQUITY_CONFIGMAP_YML=ubiquity-configmap.yml
    SCBE_CRED_YML=scbe-credentials.yml
    UBIQUITY_DB_CRED_YML=ubiquity-db-credentials.yml
    FIRST_STORAGECLASS_YML=yamls/storage-class.yml
    PVCS_USES_STORAGECLASS_YML="ubiquiyt-db-pvc.yml sanity_yamls/sanity-pvc.yml"
    declare -A KEY_FILE_DICT
    KEY_FILE_DICT['SCBE_MANAGEMENT_IP_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_MANAGEMENT_PORT_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_DEFAULT_SERVICE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['UBIQUITY_INSTANCE_NAME_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['DEFAULT_FSTYPE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['LOG_LEVEL_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SSL_MODE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SKIP_RESCAN_ISCSI_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_USERNAME_VALUE']="${SCBE_CRED_YML}"
    KEY_FILE_DICT['SCBE_PASSWORD_VALUE']="${SCBE_CRED_YML}"
    KEY_FILE_DICT['UBIQUITY_DB_USERNAME_VALUE']="${UBIQUITY_DB_CRED_YML}"
    KEY_FILE_DICT['UBIQUITY_DB_PASSWORD_VALUE']="${UBIQUITY_DB_CRED_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_NAME_VALUE']="${FIRST_STORAGECLASS_YML} ${PVCS_USES_STORAGECLASS_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_PROFILE_VALUE']="${FIRST_STORAGECLASS_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_FSTYPE_VALUE']="${FIRST_STORAGECLASS_YML}"


   base64_placeholders="UBIQUITY_DB_USERNAME_VALUE UBIQUITY_DB_PASSWORD_VALUE UBIQUITY_DB_NAME_VALUE SCBE_USERNAME_VALUE SCBE_PASSWORD_VALUE"
   was_updated="false" # if nothing to update then exit with error

   read -p "Updating ymls with placeholders from ${CONFIG_SED_FILE} file. Are you sure (y/n): " yn
   if [ "$yn" != "y" ]; then
     echo "Skip updating the ymls with placeholder."
     return
   fi

   ssl_mode=""
   # Loop over the config file and do the replacements
   for line in `cat ${CONFIG_SED_FILE} | grep -v "^\s*#"`; do
      placeholder=`echo "$line" | awk -F= '{print $1}'`
      value=`echo "$line" | awk -F= '{print $2}'`

      files_to_update=`grep ${placeholder} "$YML_DIR"/*.yml "$SANITY_YML_DIR"/*.yml "$scripts"/*.yml | awk -F: '{printf $1" "}'`
      if [ -n "$files_to_update" ]; then
         echo "$base64_placeholders" | grep "$placeholder" >/dev/null && rc=$?  || rc=$?
         replace_with_base64=""
         if [ $rc -eq 0 ]; then
            value="`echo -n $value | base64`"
            replace_with_base64="(base64) "
         fi
         printf "Update ${replace_with_base64}placeholder [%-30s] in files : $files_to_update \n" $placeholder
         sed -i "s|${placeholder}|${value}|g" $files_to_update
         [ "$placeholder" = "SSL_MODE_VALUE" ] && ssl_mode="$value" || :
         was_updated="true"
      else
         files_related="${KEY_FILE_DICT[$placeholder]}"
         if [ -z "$files_related" ]; then
            printf "WARNING : placeholder [%-30s] was NOT found in ymls\n" "$placeholder"
         else
            printf "WARNING : placeholder [%-30s] was NOT found in ymls: $files_related \n" "$placeholder"
         fi
      fi
   done

   if [ "$was_updated" = "false" ]; then
      echo "ERROR : Nothing was updated in ymls (placeholders were NOT found in ymls)."
      echo "        Consider to update yamls manually"
      exit 2
   fi

   if [ "$ssl_mode" = "verify-full" ]; then
       ymls_to_updates="${YML_DIR}/ubiquity-k8s-provisioner-deployment.yml ${YML_DIR}/ubiquity-k8s-flex-daemonset.yml ${YML_DIR}/ubiquity-deployment.yml ${YML_DIR}/ubiquity-db-deployment.yml"

       echo "Certificates updates related:"
       echo "  SSL_MODE_VALUE=verify-full, therefor updating ymls to enable dedicated certificates."
       echo "  By enable Volumes and VolumeMounts tags for certificates in the following ymls: $ymls_to_updates"

       # this sed just removes the comments from all the certificates lines on the ymls
       sed -i 's/^# Cert #\(.*\)/\1  # Cert #/g' ${ymls_to_updates}
       echo "  Certificates updates DONE."
   fi

   echo "Finish to update yaml according to ${CONFIG_SED_FILE}"
   echo ""
}

# STEP function
function create-ubiquity-db()
{
   ########################################################################
   ##  Creates only the ubiquity-db deployment and wait for creation.
   ########################################################################

    echo "Creating ubiquity-db deployment... (Assume flex plugin was already loaded on all the nodes)"
    kubectl create --namespace $NS -f ${YML_DIR}/ubiquity-db-deployment.yml
    echo "Waiting for deployment [ubiquity-db] to be created..."
    wait_for_deployment ubiquity-db 40 5 $NS
    echo ""
    echo "\"$PRODUCT_NAME\" installation finished successfully in the Kubernetes cluster. "
    echo "           - Get status      $> ./ubiquity_cli.sh -a status -n $NS"
    echo "           - Run sanity test $> ./ubiquity_cli.sh -a sanity -n $NS"
}

# STEP function (certificates related)
function create-services()
{
    ########################################################################
    ##  Creates the secrets and configmap as ubiquity ymls expect it to be.
    ##  The function operates only if:
    ##      1. $CERT_DIR dir exist and with the right key and crt files.
    ##      2. The secrets and config are not exist.
    ########################################################################
    echo "Partially install Ubiquity - creates only the ubiquity and ubiquity-db services up."
    create_only_namespace_and_services $NS
    kubectl get $nsf svc/ubiquity svc/ubiquity-db
    echo ""
    echo "Finish to create namespace, ${UBIQUITY_SERVICE_NAME} service and ${UBIQUITY_DB_SERVICE_NAME} service"
    echo "Attention: To complete Ubiqutiy installation do :"
    echo "   Prerequisite"
    echo "     (1) Generate dedicated certificates for ubiquity, ubiquity-db and scbe. (with specific name files)"
    echo "     (2) Create secrets and configmap to store the certificates and trusted CA files, by just running :"
    echo "          $> $0 -s create-secrets-for-certificates -t <certificates-directory> -n $NS"
    echo "   Complete the installation:"
    echo "     (1)  $> $0 -s install -c <file> -n $NS"
    echo "     (2)  Manually restart kubelet service on all minions to reload the new flex driver"
    echo "     (3)  $> $0 -s create-ubiquity-db -n $NS"
    echo ""
}

# STEP function (certificates related)
function create-secrets-for-certificates()
{
    ########################################################################
    ##  Creates the secrets and configmap as ubiquity ymls expect it to be.
    ##  The function operates only if:
    ##      1. $CERT_DIR dir exist and with the right key and crt files.
    ##      2. The secrets and config are not exist.
    ########################################################################
    [ -z "$CERT_DIR" ] && { echo "Error: Missing -t <file> flag for STEP [$STEP]."; exit 4; } || :
    [ ! -d "$CERT_DIR" ] && { echo "Error: $CERT_DIR directory not found."; exit 2; }

    echo "Creating secrets [ubiquity-private-certificate and ubiquity-db-private-certificate] and configmap [ubiquity-public-certificates] based from directory $CERT_DIR"

    # Validation all cert files in the $CERT_DIR directory
    expected_cert_files="ubiquity.key ubiquity.crt ubiquity-db.key ubiquity-db.crt ubiquity-trusted-ca.crt ubiquity-db-trusted-ca.crt scbe-trusted-ca.crt "
    for certfile in $expected_cert_files; do
        if [ ! -f $CERT_DIR/$certfile ]; then
            echo "Error: Missing cert file $CERT_DIR/$certfile inside directory $CERT_DIR."
            echo "       Mandatory certificate files are : $expected_cert_files"
            exit 2
        fi
    done

    # Validation secrets and configmap not exist before creation
    kubectl get secret $nsf ubiquity-db-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-db-private-certificate]" || :
    kubectl get secret $nsf ubiquity-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-private-certificate]" || :
    kubectl get configmap $nsf ubiquity-public-certificates >/dev/null 2>&1 && already_exist "configmap [ubiquity-public-certificates]" || :

    # Creating secrets and configmap
    cd $CERT_DIR
    kubectl create secret $nsf generic ubiquity-db-private-certificate --from-file=ubiquity-db.key --from-file=ubiquity-db.crt
    kubectl create secret $nsf generic ubiquity-private-certificate --from-file=ubiquity.key --from-file=ubiquity.crt
    kubectl create configmap $nsf ubiquity-public-certificates --from-file=ubiquity-db-trusted-ca.crt=ubiquity-db-trusted-ca.crt --from-file=scbe-trusted-ca.crt=scbe-trusted-ca.crt --from-file=ubiquity-trusted-ca.crt=ubiquity-trusted-ca.crt
    cd -

    kubectl get $nsf secrets/ubiquity-db-private-certificate secrets/ubiquity-private-certificate cm/ubiquity-public-certificates
    echo ""
    echo "Finished to create secrets and configmap for Ubiquity certificates."
}



function already_exist() { echo "Error: Secret $1 is already exist. Please delete it first."; exit 2; }

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

function create_configmap_and_credentials_secrets()
{
    if ! kubectl get $nsf configmap ubiquity-configmap >/dev/null 2>&1; then
        ubiquity_service_ip=`kubectl get $nsf svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
        if [ -z "$ubiquity_service_ip" ]; then
           echo "Error: Missing ubiquity service IP. The installer cannot continue without the ubiqutiy service IP."
           echo "       Review $> kubectl get $nsf svc/ubiquity"
           exit 4
        fi
        echo "Update the UBIQUITY-IP-ADDRESS: ${ubiquity_service_ip} in the file [${YML_DIR}/../ubiquity-configmap.yml]"
        sed -i "s/UBIQUITY-IP-ADDRESS:\s*\".*\"/UBIQUITY-IP-ADDRESS: \"${ubiquity_service_ip}\"/" ${YML_DIR}/../ubiquity-configmap.yml

        # Now create the configmap
        kubectl create $nsf -f ${YML_DIR}/../ubiquity-configmap.yml
    else
       echo "ubiquity-configmap configmap already exist. (Skip creation)"
    fi
    if ! kubectl get $nsf secret scbe-credentials >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/../scbe-credentials.yml
    else
       echo "scbe-credentials secret already exist. (Skip creation)"
    fi

    if ! kubectl get $nsf secret ubiquity-db-credentials >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/../ubiquity-db-credentials.yml
    else
       echo "ubiquity-db-credentials secret already exist. (Skip creation)"
    fi
}

## MAIN ##
##########

# Variables
scripts=$(dirname $0)
YML_DIR="$scripts/yamls"
SANITY_YML_DIR="$scripts/yamls/sanity_yamls"
UTILS=$scripts/ubiquity_lib.sh
UBIQUITY_DB_PVC_NAME=ibm-ubiquity-db
K8S_CONFIGMAP_FOR_PROVISIONER=k8s-config
steps="update-ymls install create-ubiquity-db create-services create-secrets-for-certificates"

[ ! -f $UTILS ] && { echo "Error: $UTILS file not found"; exit 3; }
. $UTILS # include utils for wait function and status

# Handle flags
NS="$UBIQUITY_DEFAULT_NAMESPACE" # Set as the default namespace
to_deploy_ubiquity_db="false"
KUBECONF="" #~/.kube/config
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
      found=false
      for step_index in $steps; do
          [ "$STEP" == "$step_index" ] && found=true
      done

      [ "$found" == "false" ] && { echo "Error: step [$STEP] is not supported."; usage; }
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
[ -z "$STEP" ] && usage || :


# Execute the step function
echo "Executing STEP [$STEP]..."
$STEP