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
ETC_UBIQUITY=/etc/ubiquity  # mount point in the container for /etc/ubiquity in the host

if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
  echo "Creating the flex driver directory [$DRIVER] for the first time."
  echo "***Attention*** : Kubernetes version below 1.8 - the user must restart manually the kubelet service in order to load the new flex driver."
  mkdir "${MNT_FLEX_DRIVER_DIR}"
fi

echo "Copying the flex driver ~/$DRIVER into ${MNT_FLEX_DRIVER_DIR} directory"
cp ~/$DRIVER "${MNT_FLEX_DRIVER_DIR}/.$DRIVER"
mv -f "${MNT_FLEX_DRIVER_DIR}/.$DRIVER" "${MNT_FLEX_DRIVER_DIR}/$DRIVER"

echo "Copying the flex config file ${FLEX_CONF_PATH} to ${ETC_UBIQUITY}/${FLEX_CONF}"
FLEX_TMP="${ETC_UBIQUITY}/.${FLEX_CONF}"
cp ${FLEX_CONF_PATH} ${FLEX_TMP}
if [ -n "$UBIQUITY_USERNAME" ]; then
    echo "Update \"username\" in config file based on environment UBIQUITY_USERNAME"
    sed -i "s/^username =.*/username = \"$UBIQUITY_USERNAME\"/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PASSWORD" ]; then
    echo "Update \"password\" in config file based on environment UBIQUITY_PASSWORD"
    sed -i "s/^password =.*/password = \"$UBIQUITY_PASSWORD\"/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PLUGIN_USE_SSL" ]; then
    echo "Update \"UseSsl\" in config file based on environment UBIQUITY_PLUGIN_USE_SSL"
    sed -i "s/^UseSsl =.*/UseSsl = $UBIQUITY_PLUGIN_USE_SSL/" ${FLEX_TMP}
fi
if [ -n "$UBIQUITY_PLUGIN_SSL_MODE" ]; then
    echo "Update \"SslMode\" in config file based on environment UBIQUITY_PLUGIN_SSL_MODE"
    sed -i "s/^SslMode =.*/SslMode = \"$UBIQUITY_PLUGIN_SSL_MODE\"/" ${FLEX_TMP}

   # Note: SslMode in the config file is by default verify-full
fi

if [ -n "$LOG_LEVEL" ]; then
    echo "Update \"logLevel\" in config file based on environment LOG_LEVEL"
    sed -i "s/^logLevel =.*/logLevel = \"$LOG_LEVEL\"/" ${FLEX_TMP}
fi

# Now ubiquity config file is ready with all the updates.
mv -f ${FLEX_TMP} ${ETC_UBIQUITY}/${FLEX_CONF}

if [ -n "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
   if [ -f "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
       echo "Copy the ubiquity public certificate $UBIQUITY_PLUGIN_VERIFY_CA to the host ${ETC_UBIQUITY}"
       cp $UBIQUITY_PLUGIN_VERIFY_CA ${ETC_UBIQUITY}
   else
       echo "The ubiquity public certificate $UBIQUITY_PLUGIN_VERIFY_CA file does not exist, so cannot copy it to ${ETC_UBIQUITY}"
   fi
else
       echo "The ubiquity public certificate ENV UBIQUITY_PLUGIN_VERIFY_CA is empty, so cannot copy the certificate to ${ETC_UBIQUITY}"
fi


# Run a tail -f on the flex log file (which locate on the host), so it will be visible by running kubectl logs <flex POD>
tail -F /var/tmp/ubiquity-k8s-flex.log &
echo "Finished to copy the flex driver [$DRIVER] and a config file [${FLEX_CONF}]"
while : ; do
  # TODO change to the new location later on : /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/ubiquity-k8s-flex.log &
  sleep 86400 # every 24 hours
  echo "Start ubiquity logrotate"
  /usr/sbin/logrotate /etc/logrotate.d/ubiquity_logrotate
  echo "Finish ubiquity logrotate"
done
