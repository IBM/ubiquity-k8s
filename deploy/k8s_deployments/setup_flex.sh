#!/bin/sh

###########################################################################
# Description:
# The setup_flex.sh responsible for:
# 1. Deploy flex driver & config file & trusted ca file(if exist) from the container into the host path
#    /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
# 2. Run tail -f on the flex log file, so it will be visible via kubectl logs <flex Pod>
# 3. Start infinite loop with logrotate every 24 hours on the host flex log file
#    /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/ubiquity-k8s-flex.log
#    The rotation policy is based on /etc/logrotate.d/ubiquity_logrotate
###########################################################################

set -o errexit
set -o pipefail

VENDOR=ibm
DRIVER=ubiquity-k8s-flex
DRIVER_DIR=${VENDOR}"~"${DRIVER}
MNT_FLEX=/mnt/flex  # Assume the host-path to the kubelet-plugins directory is mounted here
MNT_FLEX_DRIVER_DIR=${MNT_FLEX}/${DRIVER_DIR}
FLEX_CONF=${DRIVER}.conf
HOST_K8S_PLUGIN_DIR=/usr/libexec/kubernetes/kubelet-plugins/volume/exec

echo "Starting $DRIVER Pod [`date`]"
# Create ibm flex directory and copy the flex binary
# ------------------------------------------------------
if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
  echo "Creating the flex driver directory [$DRIVER] for the first time."
  echo "***Attention*** : If you are running on a Kubernetes version which is lower then 1.8, a restart to the kubelet service is required to take affect."
  mkdir "${MNT_FLEX_DRIVER_DIR}"
fi
echo "Copying the flex driver ~/$DRIVER into ${MNT_FLEX_DRIVER_DIR} directory."
cp ~/$DRIVER "${MNT_FLEX_DRIVER_DIR}/.$DRIVER"
mv -f "${MNT_FLEX_DRIVER_DIR}/.$DRIVER" "${MNT_FLEX_DRIVER_DIR}/$DRIVER"

# Prepare and copy the flex config file
# -----------------------------------------
ENV_LIST_FOR_FLEX_CONFIG="LOG_LEVEL, UBIQUITY_USERNAME, UBIQUITY_PASSWORD, UBIQUITY_IP_ADDRESS, SKIP_RESCAN_ISCSI, UBIQUITY_PLUGIN_USE_SSL, UBIQUITY_PLUGIN_SSL_MODE, UBIQUITY_PORT"
echo "Generating the flex config file [${MNT_FLEX_DRIVER_DIR}/${FLEX_CONF}] from environment variables: $ENV_LIST_FOR_FLEX_CONFIG"
FLEX_TMP="${MNT_FLEX_DRIVER_DIR}/.${FLEX_CONF}"

function missing_env() { echo "Error: missing environment variable $1"; exit 1; }

[ -z "$LOG_LEVEL" ] && LOG_LEVEL=info || :
[ -z "$UBIQUITY_USERNAME" ] && missing_env UBIQUITY_USERNAME || :
[ -z "$UBIQUITY_PASSWORD" ] && missing_env UBIQUITY_PASSWORD || :
[ -z "$UBIQUITY_IP_ADDRESS" ] && missing_env UBIQUITY_IP_ADDRESS || :
[ -z "$SKIP_RESCAN_ISCSI" ] && SKIP_RESCAN_ISCSI=false || :
[ -z "$UBIQUITY_PLUGIN_USE_SSL" ] && UBIQUITY_PLUGIN_USE_SSL=true || :
[ -z "$UBIQUITY_PLUGIN_SSL_MODE" ] && UBIQUITY_PLUGIN_SSL_MODE="verify-full" || :
[ -z "$UBIQUITY_PORT" ] && UBIQUITY_PORT=9999 || :

cat > $FLEX_TMP << EOF
# This file was generated automatically by the $DRIVER Pod.

logPath = "${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}"
backends = ["scbe"]
logLevel = "$LOG_LEVEL"

[UbiquityServer]
address = "$UBIQUITY_IP_ADDRESS"
port = $UBIQUITY_PORT

[CredentialInfo]
username = "$UBIQUITY_USERNAME"
password = "$UBIQUITY_PASSWORD"

[ScbeRemoteConfig]
SkipRescanISCSI = $SKIP_RESCAN_ISCSI

[SslConfig]
UseSsl = $UBIQUITY_PLUGIN_USE_SSL
SslMode = "$UBIQUITY_PLUGIN_SSL_MODE"
VerifyCa = "${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/ubiquity-trusted-ca.crt"
EOF

# Now ubiquity config file is ready with all the updates.
mv -f ${FLEX_TMP} ${MNT_FLEX_DRIVER_DIR}/${FLEX_CONF}


#  Handle verify CA certificate
# ---------------------------------
if [ -n "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
   if [ -f "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
       echo "Copy the ubiquity public certificate $UBIQUITY_PLUGIN_VERIFY_CA to the host ${MNT_FLEX_DRIVER_DIR}."
       cp $UBIQUITY_PLUGIN_VERIFY_CA ${MNT_FLEX_DRIVER_DIR}
   else
       echo "Attention: The ubiquity server certificate will not be verified. ($UBIQUITY_PLUGIN_VERIFY_CA file does not exist)"
   fi
else
       echo "Attention: The ubiquity server certificate will not be verified. (UBIQUITY_PLUGIN_VERIFY_CA environmnet variable does not exist)"
fi


echo "Finished to deploy the flex driver [$DRIVER], config file and its certificate into the host path ${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}"
echo ""
echo ""
echo "This Pod will handle log rotation for the <flex log> on the host [${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/${DRIVER}.log]"
echo "Running in the background tail -F <flex log>, so the log will be visible though kubectl logs <flex POD>"
echo "[`date`] Start to run in background #>"
echo "tail -F ${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/${DRIVER}.log"
echo "-----------------------------------------------"
tail -F ${MNT_FLEX_DRIVER_DIR}/ubiquity-k8s-flex.log &

while : ; do
  sleep 86400 # every 24 hours
  /usr/sbin/logrotate /etc/logrotate.d/ubiquity_logrotate
done
