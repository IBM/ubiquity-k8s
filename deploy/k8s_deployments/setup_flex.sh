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


echo "Copying the flex config file ${FLEX_CONF_PATH} to /etc/ubiquity/${FLEX_CONF}"
cp ${FLEX_CONF_PATH} /etc/ubiquity/.${FLEX_CONF}
mv -f /etc/ubiquity/.${FLEX_CONF} /etc/ubiquity/${FLEX_CONF}


if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
  echo "Creating the flex driver directory [$DRIVER] for the first time."
  echo "***Attention*** : Kubernetes version below 1.8 - the user must restart manually the kubelet service in order to load the new flex driver."
  mkdir "${MNT_FLEX_DRIVER_DIR}"
fi

echo "Copying the flex driver ~/$DRIVER into ${MNT_FLEX_DRIVER_DIR} directory"
cp ~/$DRIVER "${MNT_FLEX_DRIVER_DIR}/.$DRIVER"
mv -f "${MNT_FLEX_DRIVER_DIR}/.$DRIVER" "${MNT_FLEX_DRIVER_DIR}/$DRIVER"

echo "Finished to copy the flex driver [$DRIVER] and its config file [${FLEX_CONF}]"
while : ; do
  sleep 3600
done
