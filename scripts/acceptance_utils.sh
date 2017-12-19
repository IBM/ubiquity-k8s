#!/bin/bash -ex

############################################
# Utils for acceptance tests
############################################

NO_RESOURCES_STR="No resources found."
PVC_GOOD_STATUS=Bound
UBIQUITY_DEFAULT_NAMESPACE="ubiquity"
UBIQUITY_SERVICE_NAME="ubiquity"
UBIQUITY_DB_SERVICE_NAME="ubiquity-db"
PRODUCT_NAME="IBM Storage Enabler for Containers"
EXIT_WAIT_TIMEOUT_MESSAGE="Error: Script exits due to wait timeout."
SCBE_CRED_YML=scbe-credentials-secret.yml
UBIQUITY_DB_CRED_YML=ubiquity-db-credentials-secret.yml
UBIQUITY_DEPLOY_YML=ubiquity-deployment.yml
UBIQUITY_DB_DEPLOY_YML=ubiquity-db-deployment.yml
UBIQUITY_PROVISIONER_DEPLOY_YML=ubiquity-k8s-provisioner-deployment.yml
UBIQUITY_FLEX_DAEMONSET_YML=ubiquity-k8s-flex-daemonset.yml

# example : wait_for_item pvc pvc1 Bound 10 1 # wait 10 seconds till timeout
function wait_for_item()
{
  item_type=$1
  item_name=$2
  item_wanted_status=$3
  retries=$4
  max_retries=$4
  delay=$5
  while true; do
      status=`kubectl get ${item_type} ${item_name} --no-headers -o custom-columns=Status:.status.phase`
      if [ "$status" = "$item_wanted_status" ]; then
         echo "${item_type} [${item_name}] status [$status] as expected (after $(($max_retries - $retries))/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "Error: Status of item $item_name was not reached the status ${item_wanted_status}. Exiting."
             echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
             exit 2
         else
            echo "${item_type} [${item_name}] status is [$status] while expected status is [${item_wanted_status}]. sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}]"
            retries=$(($retries - 1))
            sleep $delay;
         fi
      fi
  done
}

# wait_for_item_to_delete pvc scbe-accept-voly 10 1
function wait_for_item_to_delete()
{
  item_type=$1
  item_name=$2
  retries=$3
  max_retries=$3
  delay=$4
  fail="$5"
  [ -z "$fail" ] && fail=true
  while true; do
      kubectl get ${item_type} ${item_name} && rc=$? || rc=$?
      if [ $rc -ne 0 ]; then
         echo "${item_type} [${item_name}] was deleted (after $(($max_retries - $retries))/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "${item_type} named [${item_name}] still exist after all ${max_retries} retries. exit."
             [ "$fail" = "true" ] && exit 2 || { echo "Ignore wait timeout for item ${item_type} ${item_name}. Move on."; return; }
         else
            echo "${item_type} [${item_name}] still exists. sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}]"
            retries=$(($retries - 1))
            sleep $delay;
         fi
      fi
  done
}

# TODO : need to add another wait function for container status (which is inside managed object - POD)
function add_yaml_delimiter()
{
    YAML_DELIMITER='---'
    printf "\n\n%s\n" "$YAML_DELIMITER" >> $1
}

function stepinc() { S=$(($S + 1)); }

function get_generation() {
  get_deployment_jsonpath '{.metadata.generation}' $1
}

function get_observed_generation() {
  get_deployment_jsonpath '{.status.observedGeneration}' $1
}

function get_replicas() {
  get_deployment_jsonpath '{.spec.replicas}' $1
}

function get_available_replicas() {
  get_deployment_jsonpath '{.status.availableReplicas}' $1
}

function get_deployment_jsonpath() {
  local _jsonpath="$1"

  kubectl get deployment "$2" -o "jsonpath=${_jsonpath}"
}

# wait_for_deployment ubiquity-db 3 10
function wait_for_deployment(){
  item_type=deployment
  item_name=$1
  retries=$2
  max_retries=$2
  delay=$3

    while ! kubectl get deployment $item_name > /dev/null 2>&1; do
       if [ "$retries" -eq 0 ]; then
          echo "Error: ${item_type} [${item_name}] does not exist, after all ${max_retries} retries. Exiting."
          echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
          exit 2
      else
          echo "${item_type} [${item_name}] does not exist, sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}] "
          retries=$(($retries - 1))
          sleep $delay;
      fi
    done

    generation=$(get_generation $item_name)
    while [[ $(get_observed_generation $item_name) -lt ${generation} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "Error: ${item_type} [${item_name}] generation number is [$(get_observed_generation deployment $item_name $ns)] while expected value is [${generation}], after all ${max_retries} retries. Exiting."
          echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
          exit 2
      else
          echo "${item_type} [${item_name}] generation number is [$(get_observed_generation deployment $item_name $ns)] while expected value is [${generation}], sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}] "
          retries=$(($retries - 1))
          sleep $delay;
      fi
    done
    echo "${item_type} [${item_name}] reached the expected generation [${generation}]"

    replicas="$(get_replicas $item_name)"

    available=$(get_available_replicas $item_name)
    [ -z "$available" ] && available=0
    while [[ ${available} -ne ${replicas} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "Error: ${item_type} [${item_name}] available replicas are [${available}] while expected value is [${replicas}], after all ${max_retries} retries. Exiting."
          echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
          exit 2
      else
          echo "${item_type} [${item_name}] available replicas are [${available}] while expected value is [${replicas}], sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}]"
          available=$(get_available_replicas $item_name $ns)
          [ -z "$available" ] && available=0
          retries=$(($retries - 1))
          sleep $delay;
      fi
    done

    echo "${item_type} named [${item_name}] reached to expected replicas ${replicas}"
}
