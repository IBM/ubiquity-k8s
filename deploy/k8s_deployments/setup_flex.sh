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


# Run a tail -F on the flex log file (which locate on the host), so it will be visible by running kubectl logs <flex POD>
tail -F ${MNT_FLEX_DRIVER_DIR}/ubiquity-k8s-flex.log &

echo "Finished to copy the flex driver [$DRIVER] and a config file [${FLEX_CONF}]."

while : ; do
  sleep 86400 # every 24 hours
  /usr/sbin/logrotate /etc/logrotate.d/ubiquity_logrotate
done
