#!/bin/bash -e

###################################################
# Copyright 2017 IBM Corp.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
###################################################

#------------------------------------------
# Share library for un\install scripts
#------------------------------------------

NO_RESOURCES_STR="No resources found."
PVC_GOOD_STATUS=Bound

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
         echo "${item_type} [${item_name}] status [$status] as expected (after `expr $max_retries - $retries`/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "Status of item $item_name was not reached to status ${item_wanted_status}. exit."
             exit 2
         else
            echo "${item_type} [${item_name}] status [$status] != [${item_wanted_status}] wish state. sleeping [${delay} sec] before retry to check [`expr $max_retries - $retries`/${max_retries}]"
            retries=`expr $retries - 1`
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
  while true; do
      if [ -n "$regex" ]; then
        kubectl get ${item_type} | grep "${item_name}" && rc=$? || rc=$?
      else
        kubectl get ${item_type} ${item_name} && rc=$? || rc=$?
      fi
      if [ $rc -ne 0 ]; then
         echo "${item_type} [${item_name}] was deleted (after `expr $max_retries - $retries`/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "${item_type} [${item_name}] still exist after all ${max_retries} retries. exit."
             exit 2
         else
            echo "${item_type} [${item_name}] still exist. sleeping [${delay} sec] before retry to check [`expr $max_retries - $retries`/${max_retries}]"
            retries=`expr $retries - 1`
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

function stepinc() { S=`expr $S + 1`; }

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

# e.g : wait_for_deployment ubiquity-db 3 10
function wait_for_deployment(){
  item_type=deployment
  item_name=$1
  retries=$2
  max_retries=$2
  delay=$3
  echo "Waiting for deployment [${item_name}] to be created..."

    while ! kubectl get deployment $item_name > /dev/null 2>&1; do
       if [ "$retries" -eq 0 ]; then
          echo "${item_type} [${item_name}] still not exist, even after all ${max_retries} retries. exit."
          exit 2
      else
          echo "${item_type} [${item_name}] still not exist, sleeping [${delay} sec] before retry to check [`expr $max_retries - $retries`/${max_retries}] "
          retries=`expr $retries - 1`
          sleep $delay;
      fi
    done

    generation=$(get_generation $item_name)
    while [[ $(get_observed_generation $item_name) -lt ${generation} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "${item_type} [${item_name}] generation $(get_observed_generation $item_name) < ${generation}, even after all ${max_retries} retries. exit."
          exit 2
      else
          echo "${item_type} [${item_name}] generation $(get_observed_generation $item_name) < ${generation}, sleeping [${delay} sec] before retry to check [`expr $max_retries - $retries`/${max_retries}] "
          retries=`expr $retries - 1`
          sleep $delay;
      fi
    done
    echo "${item_type} [${item_name}] reached to expected generation ${generation}"

    replicas="$(get_replicas $item_name)"

    available=$(get_available_replicas $item_name)
    [ -z "$available" ] && available=0
    while [[ ${available} -ne ${replicas} ]]; do
      if [ "$retries" -eq 0 ]; then
          echo "${item_type} [${item_name}] available replica ${available} != ${replicas}, even after all ${max_retries} retries. exit."
          exit 2
      else
          available=$(get_available_replicas $item_name)
          [ -z "$available" ] && available=0
          echo "${item_type} [${item_name}] available replica ${available} != ${replicas}, sleeping [${delay} sec] before retry to check [`expr $max_retries - $retries`/${max_retries}]"
          retries=`expr $retries - 1`
          sleep $delay;
      fi
    done

    echo "${item_type} [${item_name}] reached to expected replicas ${replicas}"
}

# e.g : is_deployment_ok ubiquity-db
function is_deployment_ok(){
  # --------------------------------------------------------
  # Description : Verify if deployment is OK.
  # Return value : if deployment ok then return code is 0, else !=0
  # --------------------------------------------------------
  item_type=deployment
  item_name=$1

  kubectl get deployment $item_name >/dev/null 2>&1 || return 1 # not exist

  [[ $(get_observed_generation $item_name) -lt $(get_generation $item_name) ]] && return 2 # observed_generation not meet

  replicas="$(get_replicas $item_name)"
  available=$(get_available_replicas $item_name)
  [ -z "$available" ] && available=0
  [[ ${available} -ne ${replicas} ]] && return 3 # replicas not meet

  return 0 # deployment is OK
}