#!/bin/sh

###########################################################################
# Description:
# The setup_flex.sh responsible for:
# 1. Deploy flex driver & trusted ca file(if exist) from the container into the host path
#    /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
# 2. Run tail -f on the flex log file, so it will be visible via kubectl logs <flex Pod>
# 3. Start infinite loop every 24 hours on the host for tailing the flex log file
###########################################################################

set -o errexit
set -o pipefail

function install_flex_driver()
{
    # Copy the flex binary to ibm flex directory
    # ------------------------------------------

    echo "Copying the flex driver ~/$DRIVER into ${MNT_FLEX_DRIVER_DIR} directory."
    cp ~/$DRIVER "${MNT_FLEX_DRIVER_DIR}/.$DRIVER"
    mv -f "${MNT_FLEX_DRIVER_DIR}/.$DRIVER" "${MNT_FLEX_DRIVER_DIR}/$DRIVER"
}

function generate_flex_conf_from_envs_and_install_it()
{
    # Create ibm flex directory
    # -------------------------
    if [ ! -d "${MNT_FLEX_DRIVER_DIR}" ]; then
      echo "Creating the flex driver directory [$DRIVER] for the first time."
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

function test_flex_driver()
{
    echo "Test the flex driver by running $> ${MNT_FLEX_DRIVER_DIR}/$DRIVER testubiquity"
    flex_log=${FLEX_LOG_DIR}/ubiquity-k8s-flex.log
	err=

	for i in {1..15}
    do
        testubiquity=`${MNT_FLEX_DRIVER_DIR}/$DRIVER testubiquity 2>&1`
        if echo "$testubiquity" | grep '"status":"Success"' >/dev/null; then
           echo "$testubiquity"
           echo "Flex test passed Ok"
           return 0
        else
            err="$testubiquity"
            echo "Flex test failed with error: $testubiquity"
            echo "Wait 2 seconds before $i retries of the flex test."
        fi

        sleep 2
    done

    # Flex cli is not working, so print latest logs and exit with error.

    if [ -f "$flex_log" ]; then
        echo "Error: Flex test was failed."
        echo "tail the flex log file $flex_log"
        echo "-----------------------[ Start view flex log ] ------------"
        tail -40 $flex_log || :
        echo "-----------------------[ End of flex log ] ------------"
    fi
    echo ""
    echo "Flex test failed with the following error:"
    echo "$err"
    echo "Error: Flex test was failed - Please check ubiquity_configmap parameters."
    exit 4
}

function install_flex_trusted_ca()
{
    #  Handle verify CA certificate
    # ---------------------------------
    if [ -n "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
       if [ -f "$UBIQUITY_PLUGIN_VERIFY_CA" ]; then
           echo "Copy the ubiquity public certificate $UBIQUITY_PLUGIN_VERIFY_CA to the host ${MNT_FLEX_DRIVER_DIR}."
           cp $UBIQUITY_PLUGIN_VERIFY_CA ${MNT_FLEX_DRIVER_DIR}
       else
           echo "*Attention*: The ubiquity server certificate will not be verified. ($UBIQUITY_PLUGIN_VERIFY_CA file does not exist)"
       fi
    else
           echo "*Attention*: The ubiquity server certificate will not be verified. (UBIQUITY_PLUGIN_VERIFY_CA environmnet variable does not exist)"
    fi
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

if [ $# -eq 1 -a $1 == "--generate_flex_conf" ]; then
    echo "Starting generating flex config file..."
    generate_flex_conf_from_envs_and_install_it

    echo "Finished to generate flex config file"
    echo ""
else
    echo "[`date`]"
    echo "Starting $DRIVER Pod..."

    install_flex_driver
    install_flex_trusted_ca

    echo "Finished to deploy the flex driver [$DRIVER], config file and its certificate into the host path ${HOST_K8S_PLUGIN_DIR}/${DRIVER_DIR}"
    echo ""

    test_flex_driver

    echo ""
    echo "This Pod will handle log rotation for the <flex log> on the host [${FLEX_LOG_DIR}/${DRIVER}.log]"
    echo "Running in the background tail -F <flex log>, so the log will be visible though kubectl logs <flex POD>"
    echo "[`date`] Start to run in background #>"
    echo "tail -F ${FLEX_LOG_DIR}/${DRIVER}.log"
    echo "-----------------------------------------------"
    tail -F ${FLEX_LOG_DIR}/${DRIVER}.log &

    while : ; do
        sleep 86400 # every 24 hours
    done
fi


