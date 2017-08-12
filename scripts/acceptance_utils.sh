#!/bin/bash -ex

############################################
# Utils for acceptance tests
############################################

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
         echo "${item_type} named [${item_name}] status [$status] as expected (after `expr $max_retries - $retries`/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "Status of item $item_name was not reached to status ${item_wanted_status}. exit."
             exit 2
         else
            echo "${item_type} named [${item_name}] status [$status] \!= [${item_wanted_status}] wish state. sleeping [$delay] before retry [`expr $max_retries - $retries`/${max_retries}]"
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
  while true; do
      kubectl get ${item_type} ${item_name} && rc=$? || rc=$?
      if [ $rc -ne 0 ]; then
         echo "${item_type} named [${item_name}] was deleted (after `expr $max_retries - $retries`/${max_retries} tries)"
         return
      else
         if [ "$retries" -eq 0 ]; then
             echo "${item_type} named [${item_name}] still exist after all ${max_retries} retries. exit."
             exit 2
         else
            echo "${item_type} named [${item_name}] still exist. sleeping [$delay] before retry [`expr $max_retries - $retries`/${max_retries}]"
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

