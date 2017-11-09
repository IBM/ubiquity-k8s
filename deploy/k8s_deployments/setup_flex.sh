#!/bin/sh

set -o errexit
set -o pipefail

VENDOR=ibm
DRIVER=ubiquity-k8s-flex
DRIVER_DIR=${VENDOR}"~"${DRIVER}
MNT_FLEX=/mnt/flex  # Assume the host-path to the kubelet-plugins directory is mounted here
MNT_FLEX_DRIVER_DIR=${MNT_FLEX}/${DRIVER_DIR}
FLEX_CONF=${DRIVER}.conf
FLEX_CONF_PATH=/mnt/ubiquity-k8s-flex-conf/${FLEX_CONF}  # The conf file to copy to the host
HOST_K8S_PLUGIN_DIR=/usr/libexec/kubernetes/kubelet-plugins/volume/exec

if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
  echo "Creating the flex driver directory [$DRIVER] for the first time."
  echo "***Attention*** : If you are running on a Kubernetes version which is lower then 1.8, a restart to the kubelet service is required to take affect."
  mkdir "${MNT_FLEX_DRIVER_DIR}"
fi

echo "Copying the flex driver ~/$DRIVER into ${MNT_FLEX_DRIVER_DIR} directory."
cp ~/$DRIVER "${MNT_FLEX_DRIVER_DIR}/.$DRIVER"
mv -f "${MNT_FLEX_DRIVER_DIR}/.$DRIVER" "${MNT_FLEX_DRIVER_DIR}/$DRIVER"

echo "Copying the flex config file ${FLEX_CONF_PATH} to ${MNT_FLEX_DRIVER_DIR}/${FLEX_CONF} ."
FLEX_TMP="${MNT_FLEX_DRIVER_DIR}/.${FLEX_CONF}"
cp ${FLEX_CONF_PATH} ${FLEX_TMP}
if [ -n "$UBIQUITY_USERNAME" ]; then
    echo "Updating \"username\" in config file based on environment variable UBIQUITY_USERNAME."
    sed -i "s/^username =.*/username = \"$UBIQUITY_USERNAME\"/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PASSWORD" ]; then
    echo "Updating \"password\" in config file based on environment variable UBIQUITY_PASSWORD."
    sed -i "s/^password =.*/password = \"$UBIQUITY_PASSWORD\"/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PLUGIN_USE_SSL" ]; then
    echo "Updating \"UseSsl\" in config file based on environment variable UBIQUITY_PLUGIN_USE_SSL."
    sed -i "s/^UseSsl =.*/UseSsl = $UBIQUITY_PLUGIN_USE_SSL/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PLUGIN_SSL_MODE" ]; then
    echo "Updating \"SslMode\" in config file based on environment variable UBIQUITY_PLUGIN_SSL_MODE."
    sed -i "s/^SslMode =.*/SslMode = \"$UBIQUITY_PLUGIN_SSL_MODE\"/" ${FLEX_TMP}

   # Note: SslMode in the config file is by default verify-full
fi

if [ -n "$LOG_LEVEL" ]; then
    echo "Updating \"logLevel\" in config file based on environment variable LOG_LEVEL."
    sed -i "s/^logLevel =.*/logLevel = \"$LOG_LEVEL\"/" ${FLEX_TMP}
fi

# Now ubiquity config file is ready with all the updates.
mv -f ${FLEX_TMP} ${MNT_FLEX_DRIVER_DIR}/${FLEX_CONF}

if [ -n "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
   if [ -f "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
       echo "Copy the ubiquity public certificate $UBIQUITY_PLUGIN_VERIFY_CA to the host ${MNT_FLEX_DRIVER_DIR}."
       cp $UBIQUITY_PLUGIN_VERIFY_CA ${MNT_FLEX_DRIVER_DIR}
   else
       echo "The ubiquity server certificate will not be verified. ($UBIQUITY_PLUGIN_VERIFY_CA file does not exist)"
   fi
else
       echo "The ubiquity server certificate will not be verified. (UBIQUITY_PLUGIN_VERIFY_CA environmnet variable does not exist)"
fi

echo "Finished to deploy the flex driver [$DRIVER], config file and its certificate into the host path ${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}"
echo ""
echo ""
echo "This Pod will handle automatically the log rotation of the flex log file on the host [${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/${DRIVER}.log]"
echo "Running in the background the command tail -F <flex log>, so the flex log will be visible though kubectl logs <flex POD>"
echo "[`date`] Start to run in background #> tail -F ${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/${DRIVER}.log"
echo "-----------------------------------------------"
tail -F ${MNT_FLEX_DRIVER_DIR}/ubiquity-k8s-flex.log &


while : ; do
  sleep 86400 # every 24 hours
  /usr/sbin/logrotate /etc/logrotate.d/ubiquity_logrotate
done
