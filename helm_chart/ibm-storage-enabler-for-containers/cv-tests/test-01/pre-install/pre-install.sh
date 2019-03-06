#!/usr/bin/env bash
#
# Pre-install script REQUIRED ONLY IF additional setup is required prior to
# helm install for this test path.  
#
# For example, if PersistantVolumes (PVs) are required for chart installation 
# they will need to be created prior to helm install.
#
# Parameters : 
#   -c <chartReleaseName>, the name of the release used to install the helm chart
#
# Pre-req environment: authenticated to cluster & kubectl cli install / setup complete

# Exit when failures occur (including unset variables)
set -o errexit
set -o nounset
set -o pipefail


# Verify pre-req environment
command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }
command -v helm > /dev/null 2>&1 || { echo "helm pre-req is missing."; exit 1; }

# Create pre-requisite components
# For example, create pre-requisite PV/PVCs using yaml definition in current directory
[[ $(dirname $0 | cut -c1) = '/' ]] && preinstallDir=$(dirname $0)/ || preinstallDir=${PWD}/$(dirname $0)/

# Process parameters notify of any unexpected
while test $# -gt 0; do
	[[ $1 =~ ^-c|--chartrelease$ ]] && { chartRelease="$2"; shift 2; continue; };
    echo "Parameter not recognized: $1, ignored"
    shift
done
: "${chartRelease:="default"}"


echo "Helm Version"
helm version

echo "VALUES FILE"
cat $preinstallDir/../values.yaml

echo "Test render the templates"
helm template $CV_TEST_CHART_DIR -n $chartRelease -f $preinstallDir/../values.yaml
