#!/bin/bash -e

#*******************************************************************************
#  Copyright 2017 IBM Corp.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#  http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#*******************************************************************************

# -------------------------------------------------------------------------
# "IBM Storage Enabler for Containers installation script Use this script to install the following components:
#       "IBM Storage Enabler for Containers"
#       "IBM Storage Dynamic Provisioner for Kubernetes"
#       "IBM Storage FlexVolume for Kubernetes"
#
# Running the script
# ==================
# Before you begin
#  1. Manually configure the relevant parameters in the ubiquity_installer.conf file.
#  2. If using the verify-full SSL mode (SSL_MODE=verify-full), perform the following:
#       2.1. Obtain the IP address of the IBM Storage Enabler for Containers (ubiquity and ubiquity-db services), which is needed for the SSL certificate generation, then run this command: $> ./ubiquity_installer.sh -s create-services.
#       2.2. Manually create dedicated SSL certificates for IBM Storage Enabler for Containers server, its database container and SCBE. This procedure is out of scope of this script.
#       2.3. Generate Kubernetes secrets for the certificates, created in the previous step. Run this command: $> ./ubiquity_installer.sh -s create-secrets-for-certificates -t <certificates-directory>.
#
#  3. Update the relevant parameters in the yml files by running this command:
#     $> ./ubiquity_installer.sh -s update-ymls -c ubiquity_installer.conf.
#
# Installation
#  1. Install IBM Storage Enabler for Containers without its database by running this command: 
#     $> ./ubiquity_installer.sh -s install
#
#  2. Manually restart the kubelet on all the Kubernetes nodes.
#
#  3. Install the IBM Storage Enabler for Containers database(ubiquity-db) by running this command: 
#     $> ./ubiquity_installer.sh -s create-ubiquity-db.
#
# -------------------------------------------------------------------------

function usage()
{
  cmd=`basename $0`
  cat << EOF
USAGE   $cmd -s <STEP> <FLAGS>
  -s <STEP>:
    -s update-ymls -c <file>
        Replace the placeholders from -c <file> in the relevant yml files.
        Flag -c <ubiquity-config-file> is mandatory for this step
    -s install [-n <namespace>]
        Installs all $PRODUCT_NAME components in orderly fashion (except for ubiquity-db).
        Flag -n <namespace>. By default, it is \"ubiquity\" namespace.
    -s create-ubiquity-db [-n <namespace>]
        Creates the ubiquity-db deployment, waiting for its creation.
        Use this option after finishing the installation and manually restarting the kubelets on the nodes.

    Steps required for SSL_MODE=verify-full:
    -s create-services [-n <namespace>]
        Creates the $PRODUCT_NAME namespace(ubiquity) and generates two Kubernetes services that provide the DNS/IP address combinations for the server and database containers.
        Flag -n <namespace>. By default, it is \"ubiquity\" namespace.
    -s create-secrets-for-certificates -t <certificates-directory>  [-n <namespace>]
        Creates secrets and configmap for the $PRODUCT_NAME certificates:
            Secrets ubiquity-private-certificate and ubiquity-db-private-certificate.
            Configmap ubiquity-public-certificates.
        Flag -t <certificates-directory> that contains all the expected certificate files.
 -h : Displays user help
EOF
  exit 1
}


# STEP function
function install()
{
    ########################################################################
    # Installs all components of IBM Storage Enabler for Containers in orderly fashion. 
    # Installation is declared successful, when the Storage Enable for Container instance is created.
    # 
    # The install step creates the following components::
    #   1. Namespace    "ubiquity"                (skip, if already exists)
    #      ServiceAccount, ClusterRoles and ClusterRolesBinding   (skip, if already exists)
    #   2. Service(clusterIP type) "ubiquity"     (skip, if already exists)
    #   3. Service(clusterIP type) "ubiquity-db"  (skip, if already exists)
    #   4. ConfigMap    "ubiquity-configmap"      (skip, if already exists)
    #   5. Secret       "scbe-credentials"        (skip, if already exists)
    #   6. Secret       "ubiquity-db-credentials" (skip, if already exists)
    #   7. Deployment   "ubiquity"
    #   8. Deployment   "ubiquity-k8s-provisioner"
    #   9. StorageClass <name given by user>      (skip, if already exists)
    #   10. PVC          "ibm-ubiquity-db"
    #   11. DaemonSet    "ubiquity-k8s-flex"
    ########################################################################

    echo "Starting installation  \"$PRODUCT_NAME\"..."
    echo "Installing on the namespace  [$NS]."

    create_only_namespace_and_services
    create_configmap_and_credentials_secrets

    kubectl create $nsf -f ${YML_DIR}/${UBIQUITY_DEPLOY_YML}
    wait_for_deployment ubiquity 20 5 $NS

    kubectl create $nsf -f ${YML_DIR}/${UBIQUITY_PROVISIONER_DEPLOY_YML}
    wait_for_deployment ubiquity-k8s-provisioner 20 5 $NS

    # Create storage class and PVC, then wait for PVC and PV creation
    if ! kubectl get $nsf -f ${YML_DIR}/storage-class.yml > /dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/storage-class.yml
    else
        echo "Skipping the creation of ${YML_DIR}/storage-class.yml storage class, because it already exists"
    fi
    kubectl create $nsf -f ${YML_DIR}/ubiquity-db-pvc.yml
    echo "Waiting for ${UBIQUITY_DB_PVC_NAME} PVC to be created"
    wait_for_item pvc ${UBIQUITY_DB_PVC_NAME} ${PVC_GOOD_STATUS} 60 5 $NS
    pvname=`kubectl get $nsf pvc ${UBIQUITY_DB_PVC_NAME} --no-headers -o custom-columns=name:spec.volumeName`
    echo "Waiting for ${pvname} PV to be created"
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 20 3 $NS

    echo "Deploying FlexVolume driver as a daemonset on all nodes and masters. The daemonset will use the ubiquity service IP."
    kubectl create $nsf -f ${YML_DIR}/${UBIQUITY_FLEX_DAEMONSET_YML}
    wait_for_daemonset ubiquity-k8s-flex 20 3 $NS

    daemonset_desiredNumberScheduled="$(get_daemonset_desiredNumberScheduled ubiquity-k8s-flex $NS)"
    number_of_nodes=`kubectl get nodes| awk '$2 ~/Ready/' | wc -l`
    flex_missing=false
    if [ "$daemonset_desiredNumberScheduled" != "$number_of_nodes" ]; then
        echo ""
        echo "*WARNING*: "
        echo "  The ubiquity-k8s-flex daemonset pod MUST run on each worker and master nodes in the cluster."
        echo "  But it run only on $daemonset_desiredNumberScheduled from $number_of_nodes nodes(and masters in the cluster)."
        flex_missing=true
    fi



    if [ "${to_deploy_ubiquity_db}" == "true" ]; then
        create-ubiquity-db
    else
        echo ""
        echo "\"$PRODUCT_NAME\" Installation finished, but the deployment is not ready yet."
        echo "  Perform the following: "
        [ "$flex_missing" = "true" ] && echo "     (0) Verify that ubiquity-k8s-flex daemonset pod runs on all nodes including all masters. If not, check why."
        echo "     (1) Manually restart the kubelet service on all Kubernetes nodes to reload the new FlexVolume driver."
        echo "     (2) Deploy ubiquity-db by $> $0 -s create-ubiquity-db -n $NS"
        echo "     Note : View status by $> ./ubiquity_cli.sh -a status -n $NS"
        echo ""
    fi
}

# STEP function
function update-ymls()
{
   ########################################################################  
   ##  Replaces all the placeholders in all the relevant yml files.
   ##  If there is nothing to replace, it exits with error.
   ########################################################################

   # Step validation
   [ -z "${CONFIG_SED_FILE}" ] && { echo "Error: Missing -c <file> flag for STEP [$STEP]"; exit 4; } || :
   [ ! -f "${CONFIG_SED_FILE}" ] && { echo "Error: ${CONFIG_SED_FILE} is not found."; exit 3; }
   which base64 > /dev/null 2>&1 || { echo "Error: Base64 command is not found in PATH. Failed to update yml files with base64 secret."; exit 2; }
   which egrep > /dev/null 2>&1 || { echo "Error: egrep command is not found in PATH. Failed to update ymls with base64 secret."; exit 2; }

   # Validate key=value file format and there are no missing VALUE
   egrep -v "^\s*#|^\s*$" ${CONFIG_SED_FILE} |  grep -v "^.\+=.\+$" && { echo "Error: ${CONFIG_SED_FILE} format must have only key=value lines. The lines above are badly formatted."; exit 2; } || :
   grep "=VALUE$" ${CONFIG_SED_FILE} && { echo "Error: You must fill in the VALUE in ${CONFIG_SED_FILE} file."; exit 2; } || :

    # Prepare map of keys inside ${CONFIG_SED_FILE} and associated yml file.
    UBIQUITY_CONFIGMAP_YML=ubiquity-configmap.yml
    FIRST_STORAGECLASS_YML=yamls/storage-class.yml
    PVCS_USES_STORAGECLASS_YML="ubiquiyt-db-pvc.yml sanity_yamls/sanity-pvc.yml"
    declare -A KEY_FILE_DICT
    KEY_FILE_DICT['UBIQUITY_IMAGE']="${UBIQUITY_DEPLOY_YML}"
    KEY_FILE_DICT['UBIQUITY_DB_IMAGE']="${UBIQUITY_DB_DEPLOY_YML}"
    KEY_FILE_DICT['UBIQUITY_K8S_PROVISIONER_IMAGE']="${UBIQUITY_PROVISIONER_DEPLOY_YML}"
    KEY_FILE_DICT['UBIQUITY_K8S_FLEX_IMAGE']="${UBIQUITY_FLEX_DAEMONSET_YML}"
    KEY_FILE_DICT['SCBE_MANAGEMENT_IP_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_MANAGEMENT_PORT_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_DEFAULT_SERVICE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['UBIQUITY_INSTANCE_NAME_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['IBM_UBIQUITY_DB_PV_NAME_VALUE']="${UBIQUITY_CONFIGMAP_YML} ${PVCS_USES_STORAGECLASS_YML}"
    KEY_FILE_DICT['DEFAULT_FSTYPE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['FLEX_LOG_DIR_VALUE']="${UBIQUITY_CONFIGMAP_YML} ${UBIQUITY_FLEX_DAEMONSET_YML}"
    KEY_FILE_DICT['LOG_LEVEL_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SSL_MODE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SKIP_RESCAN_ISCSI_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['DEFAULT_VOLUME_SIZE_VALUE']="${UBIQUITY_CONFIGMAP_YML}"
    KEY_FILE_DICT['SCBE_USERNAME_VALUE']="${SCBE_CRED_YML}"
    KEY_FILE_DICT['SCBE_PASSWORD_VALUE']="${SCBE_CRED_YML}"
    KEY_FILE_DICT['UBIQUITY_DB_USERNAME_VALUE']="${UBIQUITY_DB_CRED_YML}"
    KEY_FILE_DICT['UBIQUITY_DB_PASSWORD_VALUE']="${UBIQUITY_DB_CRED_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_NAME_VALUE']="${FIRST_STORAGECLASS_YML} ${PVCS_USES_STORAGECLASS_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_PROFILE_VALUE']="${FIRST_STORAGECLASS_YML}"
    KEY_FILE_DICT['STORAGE_CLASS_FSTYPE_VALUE']="${FIRST_STORAGECLASS_YML}"


   base64_placeholders="UBIQUITY_DB_USERNAME_VALUE UBIQUITY_DB_PASSWORD_VALUE UBIQUITY_DB_NAME_VALUE SCBE_USERNAME_VALUE SCBE_PASSWORD_VALUE"
   was_updated="false" # if there is nothing to update, exit with error

   read -p "Updating yml files with placeholders from ${CONFIG_SED_FILE} file. Are you sure (y/n): " yn
   if [ "$yn" != "y" ]; then
     echo "Skip updating the yml files with placeholder."
     return
   fi

   ssl_mode=""
   # Loop over the configuration file and replace the values
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
            printf "WARNING: placeholder [%-30s] was NOT found in yml files\n" "$placeholder"
         else
            printf "WARNING: placeholder [%-30s] was NOT found in yml files: $files_related \n" "$placeholder"
         fi
      fi
   done

   if [ "$was_updated" = "false" ]; then
      echo "ERROR: Nothing was updated in yml files  (placeholders were NOT found in yml files)."
      echo "        Consider updating yml files manually"
      exit 2
   fi

   if [ "$ssl_mode" = "verify-full" ]; then
       ymls_to_updates="${YML_DIR}/${UBIQUITY_PROVISIONER_DEPLOY_YML} ${YML_DIR}/${UBIQUITY_FLEX_DAEMONSET_YML} ${YML_DIR}/${UBIQUITY_DEPLOY_YML} ${YML_DIR}/${UBIQUITY_DB_DEPLOY_YML}"

       echo "Related to certificate update:"
       echo "  SSL_MODE_VALUE=verify-full, updating yml files  to enable dedicated certificates."
       echo "  By enable Volumes and VolumeMounts tags for certificates in the following yml files: $ymls_to_updates"

       # this sed removes the comments from all the certificate lines in the yml files
       sed -i 's/^# Cert #\(.*\)/\1  # Cert #/g' ${ymls_to_updates}
       echo "  Certificate updates are completed."
   fi

   echo "Finished updating yml files according to ${CONFIG_SED_FILE}"
}

# STEP function
function create-ubiquity-db()
{
   ########################################
   ##  Creates deployment for ubiquity-db
   ########################################

    echo "Creating ubiquity-db deployment... (Assuming that the IBM Storage Kubernetes FlexVolume(ubiquity-k8s-flex) plugin is already loaded on all the nodes)"
    kubectl create --namespace $NS -f ${YML_DIR}/${UBIQUITY_DB_DEPLOY_YML}
    echo "Waiting for deployment [ubiquity-db] to be created..."
    wait_for_deployment ubiquity-db 50 5 $NS
    echo ""
    echo "\"$PRODUCT_NAME\" installation finished successfully in the Kubernetes cluster. "
    echo "           - Get status      $> ./ubiquity_cli.sh -a status -n $NS"
    echo "           - Run sanity test $> ./ubiquity_cli.sh -a sanity -n $NS"
}

# STEP function (certificates related)
function create-services()
{

    ########################################################################
    # The create-services creates the following components:
    #   1. Namespace    "ubiquity"                (skip, if already exists)
    #   2. Service(clusterIP type) "ubiquity"     (skip, if already exists)
    #   3. Service(clusterIP type) "ubiquity-db"  (skip, if already exists)
    ########################################################################
    echo "Partially installs - creates only the ubiquity and ubiquity-db services."
    create_only_namespace_and_services $NS
    kubectl get $nsf svc/ubiquity svc/ubiquity-db
    echo ""
    echo "Finished creating namespace, ${UBIQUITY_SERVICE_NAME} service and ${UBIQUITY_DB_SERVICE_NAME} service"
    echo "Attention: To complete the $PRODUCT_NAME installation with SSL_MODE=verify-full:"
    echo "   Prerequisite:"
    echo "     (1) Generate dedicated certificates for 'ubiquity', 'ubiquity-db' and SCBE, using specific file names"
    echo "     (2) Create secrets and ConfigMap to store the certificates and trusted CA files by running::"
    echo "          $> $0 -s create-secrets-for-certificates -t <certificates-directory> -n $NS"
    echo "   Complete the installation:"
    echo "     (1)  $> $0 -s install -n $NS"
    echo "     (2)  Manually restart kubelet service on all kubernetes nodes to reload the new FlexVolume driver"
    echo "     (3)  $> $0 -s create-ubiquity-db -n $NS"
    echo ""
}

# STEP function (related to certificates)
function create-secrets-for-certificates()
{
    ########################################################################
    # The create-secrets-for-certificates step creates the following components:
    #   1. secret    "ubiquity-db-private-certificate"
    #   2. secret    "ubiquity-private-certificate"
    #   3. configmap "ubiquity-public-certificates"
    #
    #  Prerequisite:
    #      1. $CERT_DIR dir exists and contains correct key and crt files.
    #      2. The secrets and ConfigMap do not exist.
    ########################################################################
    [ -z "$CERT_DIR" ] && { echo "Error: Missing -t <file> flag for STEP [$STEP]."; exit 4; } || :
    [ ! -d "$CERT_DIR" ] && { echo "Error: $CERT_DIR directory is not found."; exit 2; }

    echo "Creating secrets [ubiquity-private-certificate and ubiquity-db-private-certificate] and ConfigMap [ubiquity-public-certificates] based on files in directory $CERT_DIR"

    # Validating all certificate files in the $CERT_DIR directory
    expected_cert_files="ubiquity.key ubiquity.crt ubiquity-db.key ubiquity-db.crt ubiquity-trusted-ca.crt ubiquity-db-trusted-ca.crt scbe-trusted-ca.crt "
    for certfile in $expected_cert_files; do
        if [ ! -f $CERT_DIR/$certfile ]; then
            echo "Error: Missing certificate file $CERT_DIR/$certfile in directory $CERT_DIR."
            echo "       Mandatory certificate files are: $expected_cert_files"
            exit 2
        fi
    done

    # Veryfying that secrets and ConfigMap do not exist before starting their creation
    kubectl get secret $nsf ubiquity-db-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-db-private-certificate]" || :
    kubectl get secret $nsf ubiquity-private-certificate >/dev/null 2>&1 && already_exist "secret [ubiquity-private-certificate]" || :
    kubectl get configmap $nsf ubiquity-public-certificates >/dev/null 2>&1 && already_exist "configmap [ubiquity-public-certificates]" || :

    # Creating secrets and ConfigMap
    cd $CERT_DIR
    kubectl create secret $nsf generic ubiquity-db-private-certificate --from-file=ubiquity-db.key --from-file=ubiquity-db.crt
    kubectl create secret $nsf generic ubiquity-private-certificate --from-file=ubiquity.key --from-file=ubiquity.crt
    kubectl create configmap $nsf ubiquity-public-certificates --from-file=ubiquity-db-trusted-ca.crt=ubiquity-db-trusted-ca.crt --from-file=scbe-trusted-ca.crt=scbe-trusted-ca.crt --from-file=ubiquity-trusted-ca.crt=ubiquity-trusted-ca.crt
    cd -

    kubectl get $nsf secrets/ubiquity-db-private-certificate secrets/ubiquity-private-certificate cm/ubiquity-public-certificates
    echo ""
    echo "Finished creating secrets and ConfigMap for $PRODUCT_NAME certificates."
}



function already_exist() { echo "Error: Secret $1 already exists. Delete it first."; exit 2; }

function create_only_namespace_and_services()
{
    # This function create the ubiquity namespace, ubiquity service and ubiquity-db service, if they do not exist

    # Create ubiquity namespace
    if [ "$NS" = "$UBIQUITY_DEFAULT_NAMESPACE" ]; then
        if ! kubectl get namespace $UBIQUITY_DEFAULT_NAMESPACE >/dev/null 2>&1; then
           kubectl create -f ${YML_DIR}/ubiquity-namespace.yml
        else
           echo "$UBIQUITY_DEFAULT_NAMESPACE already exist. (Skip namespace creation)"
        fi
    fi

    create_serviceaccount_and_clusterroles

    # Creating ubiquity service
    if ! kubectl get $nsf service ${UBIQUITY_SERVICE_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-service.yml
    else
       echo "$UBIQUITY_SERVICE_NAME service already exists,skipping service creation"
    fi

    # Create ubiquity-db service
    if ! kubectl get $nsf service ${UBIQUITY_DB_SERVICE_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-db-service.yml
    else
       echo "$UBIQUITY_DB_SERVICE_NAME service already exists, skipping service creation"
    fi
}

function create_serviceaccount_and_clusterroles()
{
    # Creating ubiquity service account
    if ! kubectl get $nsf serviceaccount ${UBIQUITY_SERVICEACCOUNT_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-serviceaccount.yml
    else
       echo "${UBIQUITY_SERVICEACCOUNT_NAME} serviceaccount already exists,skipping serviceaccount creation"
    fi

    # Creating ubiquity clusterRoles
    if ! kubectl get $nsf clusterroles ${UBIQUITY_CLUSTERROLES_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-clusterroles.yml
    else
       echo "${UBIQUITY_CLUSTERROLES_NAME} clusterRoles already exists,skipping clusterRoles creation"
    fi

    # Creating ubiquity clusterRolesBindings
    if ! kubectl get $nsf clusterrolebindings ${UBIQUITY_CLUSTERROLESBINDING_NAME} >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/ubiquity-clusterrolebindings-k8s.yml
    else
       echo "${UBIQUITY_CLUSTERROLESBINDING_NAME} clusterrolebindings already exists,skipping clusterrolebindings creation"
    fi
}

function create_configmap_and_credentials_secrets()
{
    if ! kubectl get $nsf configmap ubiquity-configmap >/dev/null 2>&1; then
        ubiquity_service_ip=`kubectl get $nsf svc/ubiquity -o=custom-columns=:.spec.clusterIP | tail -1`
        if [ -z "$ubiquity_service_ip" ]; then
           echo "Error: Missing ubiquity service IP. Installation halted."
           echo "       Review $> kubectl get $nsf svc/ubiquity"
           exit 4
        fi
        echo "Update the UBIQUITY-IP-ADDRESS: ${ubiquity_service_ip} in the file [${YML_DIR}/../ubiquity-configmap.yml]"
        sed -i "s/UBIQUITY-IP-ADDRESS:\s*\".*\"/UBIQUITY-IP-ADDRESS: \"${ubiquity_service_ip}\"/" ${YML_DIR}/../ubiquity-configmap.yml

        # Creating ConfigMap
        kubectl create $nsf -f ${YML_DIR}/../ubiquity-configmap.yml
    else
       echo "ubiquity-configmap ConfigMap already exists,skipping ConfigMap creation"
    fi
    if ! kubectl get $nsf secret scbe-credentials >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/../${SCBE_CRED_YML}
    else
       echo "scbe-credentials secret already exists,skipping secret creation"
    fi

    if ! kubectl get $nsf secret ubiquity-db-credentials >/dev/null 2>&1; then
        kubectl create $nsf -f ${YML_DIR}/../${UBIQUITY_DB_CRED_YML}
    else
       echo "ubiquity-db-credentials secret already exists, skipping secret creation"
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
steps="update-ymls install create-ubiquity-db create-services create-secrets-for-certificates"

[ ! -f $UTILS ] && { echo "Error: $UTILS file is not found"; exit 3; }
. $UTILS # include utils for wait function and status

# Handle flags
NS="$UBIQUITY_DEFAULT_NAMESPACE" # Set as the default namespace
to_deploy_ubiquity_db="false"
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
which kubectl > /dev/null 2>&1 || { echo "Error: kubectl is not found in PATH"; exit 2; }
[ -z "$STEP" ] && usage || :


# Execute the step function
echo "Executing STEP [$STEP]..."
$STEP
