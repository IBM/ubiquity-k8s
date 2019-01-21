#!/bin/sh

###########################################################################
# Description:
# The generate_flex_conf.sh responsible for:
# 1. Deploy flex config file from the container into the host path
#    /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
###########################################################################

set -o errexit
set -o pipefail

function generate_flex_conf_from_envs_and_install_it()
{
    # Create ibm flex directory
    # -------------------------
    if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
      echo "Creating the flex driver directory [$DRIVER] for the first time."
      echo "***Attention*** : If you are running on a Kubernetes version which is lower then 1.8, a restart to the kubelet service is required to take affect."
      mkdir "${MNT_FLEX_DRIVER_DIR}"
    fi

    # Generate and copy the flex config file
    # --------------------------------------
    echo "Generating the flex config file(from environment variables) on the host path [${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/${FLEX_CONF}]."
    FLEX_TMP="${MNT_FLEX_DRIVER_DIR}/.${FLEX_CONF}"

    function missing_env() { echo "Error: missing environment variable $1"; exit 1; }

    # Mandatory environment variable
    # UBIQUITY_USERNAME and UBIQUITY_PASSWORD are not mandatory for Spectrum Scale hence commented
    #[ -z "$UBIQUITY_USERNAME" ] && missing_env UBIQUITY_USERNAME || :
    #[ -z "$UBIQUITY_PASSWORD" ] && missing_env UBIQUITY_PASSWORD || :

    # Other environment variable with default values
    [ -z "$FLEX_LOG_DIR" ] && FLEX_LOG_DIR=/var/log || :
    [ -z "$FLEX_LOG_ROTATE_MAXSIZE" ] && FLEX_LOG_ROTATE_MAXSIZE=50 || :
    [ -z "$LOG_LEVEL" ] && LOG_LEVEL=info || :
    [ -z "$UBIQUITY_PLUGIN_USE_SSL" ] && UBIQUITY_PLUGIN_USE_SSL=true || :
    [ -z "$UBIQUITY_PLUGIN_SSL_MODE" ] && UBIQUITY_PLUGIN_SSL_MODE="verify-full" || :
    [ -z "$UBIQUITY_PORT" ] && UBIQUITY_PORT=9999 || :
    [ -z "$UBIQUITY_BACKEND" ] && UBIQUITY_BACKEND=scbe || :

    cat > $FLEX_TMP << EOF
# This file was generated automatically by the $DRIVER Pod.

logPath = "$FLEX_LOG_DIR"
logRotateMaxSize = $FLEX_LOG_ROTATE_MAXSIZE
backends = ["$UBIQUITY_BACKEND"]
logLevel = "$LOG_LEVEL"

[UbiquityServer]
address = "0.0.0.0"
port = $UBIQUITY_PORT

[CredentialInfo]
username = "$UBIQUITY_USERNAME"
password = "$UBIQUITY_PASSWORD"

[SslConfig]
UseSsl = $UBIQUITY_PLUGIN_USE_SSL
SslMode = "$UBIQUITY_PLUGIN_SSL_MODE"
VerifyCa = "${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}/ubiquity-trusted-ca.crt"
EOF

    # Now ubiquity config file is ready with all the updates.
    mv -f ${FLEX_TMP} ${MNT_FLEX_DRIVER_DIR}/${FLEX_CONF}
}

### MAIN ###
############

VENDOR=ibm
DRIVER=ubiquity-k8s-flex
DRIVER_DIR=${VENDOR}"~"${DRIVER}
HOST_K8S_PLUGIN_DIR=/usr/libexec/kubernetes/kubelet-plugins/volume/exec   # Assume the host-path to the kubelet-plugins directory is mounted here
MNT_FLEX=${HOST_K8S_PLUGIN_DIR}
MNT_FLEX_DRIVER_DIR=${MNT_FLEX}/${DRIVER_DIR}
FLEX_CONF=${DRIVER}.conf

echo "Starting $DRIVER Pod init container..."
generate_flex_conf_from_envs_and_install_it

echo "Finished to generate flex config file"
echo ""
