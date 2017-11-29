#!/bin/bash -e

#*******************************************************************************
#  Copyright 2017 IBM Corp.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#  http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#*******************************************************************************

#------------------------------------------
# Share library for ubiquity_installer.sh and ubiquity_uninstall.sh scripts
#------------------------------------------

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



# Example: wait_for_item pvc pvc1 Bound 10 1 ubiquity   # wait 10 seconds till timeout
function wait_for_item()
{
  item_type=$1
  item_name=$2
  item_wanted_status=$3
  retries=$4
  max_retries=$4
  delay=$5
  ns=$6
  while true; do
      status=`kubectl get --namespace $ns ${item_type} ${item_name} --no-headers -o custom-columns=Status:.status.phase`
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
  regex=$5
  ns=$6
  while true; do
      if [ -n "$regex" ]; then
        kubectl get --namespace $ns ${item_type} 2>/dev/null | grep "${item_name}"  && rc=$? || rc=$?
      else
        kubectl get --namespace $ns  ${item_type} ${item_name} 2>/dev/null && rc=$? || rc=$?
        # send stderr to /dev/null to avoid printing the "Error... not found" to console.
      fi
      if [ $rc -ne 0 ]; then
         echo "${item_type} [${item_name}] was deleted (after $(($max_retries - $retries))/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "Error: ${item_type} [${item_name}] still exists after all ${max_retries} retries. Exiting."
             echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
             exit 2
         else
            echo "${item_type} [${item_name}] still exists. sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}]"
            retries=$(($retries - 1))
            sleep $delay;
         fi
      fi
  done
}



# TODO: need to add another wait function for container status (which is inside managed object - POD)
function add_yaml_delimiter()
{
    YAML_DELIMITER='---'
    printf "\n\n%s\n" "$YAML_DELIMITER" >> $1
}

function stepinc() { S=$(($S + 1)); }

function get_generation() {
  # Args : $1 object type $2 object name, $3 namespace
  get_object_jsonpath $1 '{.metadata.generation}' $2 $3
}

function get_observed_generation() {
  # Args : $1 object type $2 object name, $3 namespace
  get_object_jsonpath $1 '{.status.observedGeneration}' $2 $3
}

function get_replicas() {
  # Args : $1 object name, $2 namespace
  get_object_jsonpath deployment '{.spec.replicas}' $1 $2
}

function get_available_replicas() {
  # Args : $1 object name, $2 namespace
  get_object_jsonpath deployment '{.status.availableReplicas}' $1 $2
}

function get_daemonset_numberAvailable() {
  # Args : $1 object name, $2 namespace
  get_object_jsonpath daemonset '{.status.numberAvailable}' $1 $2
}

function get_daemonset_desiredNumberScheduled() {
  # Args : $1 object name, $2 namespace
  get_object_jsonpath daemonset '{.status.desiredNumberScheduled}' $1 $2
}

function get_object_jsonpath() {
  # Args : $1 jsonpath, $2 object name, $3 namespace
  item_type="$1"
  local _jsonpath="$2"
  item="$3"
  ns="$4"

  kubectl get --namespace $ns $item_type "$item" -o "jsonpath=${_jsonpath}"
}

# For example, wait_for_deployment ubiquity-db 3 10 ubiquity
function wait_for_deployment(){
  # Args : $1 object type, $2 object name, $3 retries,
  #        $4 max retries, $5 delay between retries, $6 namespace
  item_type=deployment
  item_name=$1
  retries=$2
  max_retries=$2
  delay=$3
  ns=$4
  echo "Waiting for deployment [${item_name}] to be in Running state..."

    while ! kubectl get --namespace $ns deployment $item_name > /dev/null 2>&1; do
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

    generation=$(get_generation deployment $item_name $ns)
    while [[ $(get_observed_generation deployment $item_name $ns) -lt ${generation} ]]; do
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

    replicas="$(get_replicas $item_name $ns)"

    available=$(get_available_replicas $item_name $ns)
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

    echo "${item_type} [${item_name}] reached the expected replicas [${replicas}]"
}


# e.g : wait_for_deployment ubiquity-db 3 10 ubiquity
function wait_for_daemonset(){
  # Args : $1 object type, $2 object name, $3 retries,
  #        $4 max retries, $5 delay between retries, $6 namespace
  item_type=daemonset
  item_name=$1
  retries=$2
  max_retries=$2
  delay=$3
  ns=$4
  echo "Waiting for daemonset [${item_name}] to be in Running state..."

    while ! kubectl get --namespace $ns daemonset $item_name > /dev/null 2>&1; do
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

    generation=$(get_generation daemonset $item_name $ns)
    while [[ $(get_observed_generation daemonset $item_name $ns) -lt ${generation} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "Error: ${item_type} [${item_name}] generation is [$(get_observed_generation daemonset $item_name $ns]) while the expected value is [${generation}], after all ${max_retries} retries. Exiting."
          echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
          exit 2
      else
          echo "${item_type} [${item_name}] generation is [$(get_observed_generation daemonset $item_name $ns)] while the expected value is [${generation}], sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}] "
          retries=$(($retries - 1))
          sleep $delay;
      fi
    done
    echo "${item_type} [${item_name}] reached the expected generation [${generation}]"

    replicas="$(get_daemonset_desiredNumberScheduled $item_name $ns)"

    available=$(get_daemonset_numberAvailable $item_name $ns)
    [ -z "$available" ] && available=0
    while [[ ${available} -ne ${replicas} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "Error: ${item_type} [${item_name}] available pods are [${available}] while expected quantaty is [${replicas}], after all ${max_retries} retries. Exiting."
          echo "$EXIT_WAIT_TIMEOUT_MESSAGE"
          exit 2
      else
          echo "${item_type} [${item_name}] available pods are [${available}] while expected quantity is [${replicas}], sleeping [${delay} sec] before retrying to check [$(($max_retries - $retries))/${max_retries}]"
          available=$(get_daemonset_numberAvailable $item_name $ns)
          [ -z "$available" ] && available=0
          retries=$(($retries - 1))
          sleep $delay;
      fi
    done

    echo "${item_type} [${item_name}] reached the expected available pods [${replicas}]"
}

# e.g : is_deployment_ok ubiquity-db ubiquity
function is_deployment_ok(){
  # --------------------------------------------------------
  # Description: Verify that deployment is OK.
  # Return value: if deployment OK, return code is 0, else !=0
  # --------------------------------------------------------
  item_type=deployment
  item_name=$1
  ns=$2

  kubectl get --namespace $ns deployment $item_name >/dev/null 2>&1 || return 1 # does not exist

  [[ $(get_observed_generation deployment $item_name $ns) -lt $(get_generation deployment $item_name $ns) ]] && return 2 # observed_generation not met

  replicas="$(get_replicas $item_name $ns)"
  available=$(get_available_replicas $item_name $ns)
  [ -z "$available" ] && available=0
  [[ ${available} -ne ${replicas} ]] && return 3 # replicas not met

  return 0 # deployment is OK
}