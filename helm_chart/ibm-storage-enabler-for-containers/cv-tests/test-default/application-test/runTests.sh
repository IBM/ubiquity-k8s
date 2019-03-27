#!/bin/bash
#
# runTests script REQUIRED ONLY IF additional application verification is 
# needed above and beyond helm tests.
#
# Parameters : 
#   -c <chartReleaseName>, the name of the release used to install the helm chart
#
# Pre-req environment: authenticated to cluster, kubectl cli install / setup complete, & chart installed

# Exit when failures occur (including unset variables)
set -o errexit
set -o nounset
set -o pipefail
 
# Process parameters notify of any unexpected
while test $# -gt 0; do
        [[ $1 =~ ^-c|--chartrelease$ ]] && { chartRelease="$2"; shift 2; continue; };
    echo "Parameter not recognized: $1, ignored"
    shift
done
: "${chartRelease:="default"}"



# Parameters 
# Below is the current set of parameters which are passed in to the app test script.
# The script can process or ignore the parameters
# The script can be coded to expect the parameter list below, but should not be coded such that additional parameters
# will cause the script to fail
#   -e <environment>, IP address of the environment
#   -r <release>, ie V.R.M.F-tag, the release notation associated with the environment, this will be V.R.M.F, plus an option -tag
#   -a <architecture>, the architecture of the environment 
#   -u <userid>, the admin user id for the environment
#   -p <password>, the password for accessing the environment, base64 encoded, p=`echo p_enc | base64 -d` to decode the password when using


# Verify pre-req environment
command -v kubectl > /dev/null 2>&1 || { echo "kubectl pre-req is missing."; exit 1; }

# Setup and execute application test on installation
echo "Running Ubiquity application test on $chartRelease"
output=$(kubectl get po --all-namespaces | grep "ubiquity*"  | awk '{ print $4 }' | sed -n '1p')

if [[ "$output" = "Running" ]]; then
         echo "Ubiquity pod(s) runnning ok."
else
         exit 1;
fi


#output=$(kubectl get svc | grep "*ubiquity*" |  awk '{ print $5 }')

#if [[ "$output" = "9999/TCP" ]] || [["$output" = "5432/TCP"]]; then
#         echo "Ubiquity service running ok."
#else
#         exit 1;
#fi
