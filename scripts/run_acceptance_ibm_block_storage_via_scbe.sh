#!/bin/bash -ex

############################################
# Acceptance Test for IBM Block Storage via SCBE
# Script prerequisites:
#    1. SCBE server up and running with 1 service delegated to ubiquity interface
#    2. ubiqutiy server up and running with SCBE backend configured
#    3. ubiquity-docker-plugin up and running
#    4. setup connectivity between the docker node to the related storage system of the service.
#
#   Two nodes tests :
#      In case second node provided a migration tests will take place
#      prerequisites for that is : the second node should apply to #3, #4 and has ssh keys from current node to the second node.
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

function add_yaml_delimiter()
{
    printf "\n\n%s\n" "$YAML_DELIMITER" >> $1
}

function basic_tests_on_one_node()
{
    # Description of the test :
    # -------------------------
    # The test creates and validate the following objects: SC, PVC, PV, POD with PVC.
    # The test validate that the relevant mount, multipathing was created after POD is up.
    # Then generate some IO on the volume inside the container, delete the POD and start it again
    # and validate that the data still persist.
    # -------------------------

	echo "####### ---> ${S}. Verify that no volume attached to the kube node1"
	ssh root@$node1 'df | egrep "ubiquity"' && exit 1 || :
	ssh root@$node1 'multipath -ll | grep IBM' && exit 1 || :
	ssh root@$node1 'lsblk | egrep "ubiquity" -B 1' && exit 1 || :
	kubectl get pvc 2>&1 | grep "$NO_RESOURCES_STR"
	kubectl get pv 2>&1 | grep "$NO_RESOURCES_STR"


	stepinc
	echo "####### ---> ${S}. Creating Storage class for ${profile} service"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}
    cat $yml_sc_profile
    kubectl create -f ${yml_sc_profile}
    kubectl get storageclass $profile

	echo "####### ---> ${S}. Create PVC (volume) on SCBE ${profile} service (which is on IBM FlashSystem A9000R)"
    yml_pvc=$scripts/../deploy/scbe_volume_pvc_${PVCName}.yml
    cp -f ${yml_pvc_template} ${yml_pvc}
    sed -i -e "s/PVCNAME/$PVCName/g" -e "s/SIZE/5Gi/g" -e "s/PROFILE/$profile/g" ${yml_pvc}
    cat ${yml_pvc}
    kubectl create -f ${yml_pvc}

	echo "####### ---> ${S}.1. Verify PVC and PV info status and inpect"
    wait_for_item pvc $PVCName ${PVC_GOOD_STATUS} 5 2

    pvname=`kubectl get pvc $PVCName --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 5 2
    kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname

    wwn=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname`
    kubectl get pv -o json $pvname | grep -A15 flexVolume

	echo "## ---> ${S}.2. Verify storage side : verify the volume was created on the relevant pool\service"
	echo "Skip step"
	## ssh root@gen4d-67a "xcli.py vol_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run POD [$PODName] with container ${CName} with the new volume"
    yml_pod1=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod1}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" ${yml_pod1}
    cat $yml_pod1
    kubectl create -f ${yml_pod1}
    wait_for_item pod $PODName Running 15 3


	echo "## ---> ${S}.1. Verify the volume was attached to the kubelet node $node1"
	ssh root@$node1 "df | egrep ubiquity | grep $wwn"
	ssh root@$node1 "multipath -ll | grep -i $wwn"
	ssh root@$node1 'lsblk | egrep "ubiquity|^NAME" -B 1'
	ssh root@$node1 "mount |grep $wwn| grep $fstype"

	echo "## ---> ${S}.2. Verify volume exist inside the container"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data"

	echo "## ---> ${S}.3. Verify container with the mount point"
    kubectl describe pod $PODName | grep -A1 "Volume Mounts"

	echo "## ---> ${S}.3. Verify the storage side : check volume has mapping to the host"
    echo "Skip step"
    ## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Write DATA on the volume by create a file in /data inside the container"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"

	stepinc
	echo "####### ---> ${S}. Stop the container"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 15 3

	echo "## ---> ${S}.1. Verify the volume was detached from the kubelet node"
	sleep 2 # some times mount is not refreshed immediate
	ssh root@$node1 "df | egrep ubiquity | grep $wwn" && exit 1 || :
	ssh root@$node1 "multipath -ll | grep -i $wwn" && exit 1 || :
	ssh root@$node1 'lsblk | egrep "ubiquity" -B 1' && exit 1 || :
	ssh root@$node1 "mount |grep $wwn| grep $fstype" && exit 1 || :

	echo "## ---> ${S}.2. Verify PVC and PV still exist"
    kubectl get pvc $PVCName
    kubectl get pv $pvname

	echo "## ---> ${S}.3. Verify the storage side : check volume is no longer mapped to the hos"
    echo "Skip step"
	## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run the POD again with the same volume and check the if the data remains"
    kubectl create -f ${yml_pod1}
    wait_for_item pod $PODName Running 15 3

	echo "## ---> ${S}.1. Verify that the data remains (file exist on the /data inside the container)"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"


	stepinc
	echo "####### ---> ${S}. Stop the POD"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 15 3

	stepinc
	echo "####### ---> ${S}. Remove the PVC and PV"
	kubectl delete -f ${yml_pvc}
    wait_for_item_to_delete pvc $PVCName 10 2
    wait_for_item_to_delete pv $pvname 10 2

	echo "## ---> ${S}.1. Verity the storage side : check volume is no longer exist"
    echo "Skip step"
	##  ssh root@[A9000] "xcli.py vol_list vol=u_ubiquity_instance1_$vol"

	stepinc
	echo "####### ---> ${S}. Remove the Storage Class $profile"
    kubectl delete -f ${yml_sc_profile}
    wait_for_item_to_delete storageclass $profile 5 2
}

function basic_tests_on_one_node_sc_pvc_pod_all_in_one()
{
    # Description of the test :
    # -------------------------
    # This test generates yml file with definition of many object type, such as SC, PVC and POD
    # Then create these object in one kubectl command and validate all is up and running.
    # -------------------------

	stepinc
	echo "########################### [All in one suite] ###############"
	echo "####### ---> ${S}. Prepare all in one yaml with SC, PVC, POD yml"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}

    yml_pvc=$scripts/../deploy/scbe_volume_pvc_${PVCName}.yml
    cp -f ${yml_pvc_template} ${yml_pvc}
    sed -i -e "s/PVCNAME/$PVCName/g" -e "s/SIZE/5Gi/g" -e "s/PROFILE/$profile/g" ${yml_pvc}

    yml_pod1=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod1}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" ${yml_pod1}

	ymk_sc_and_pvc_and_pod1=$scripts/../deploy/scbe_volume_with_sc_pvc_and_pod1.yml
	cat ${yml_sc_profile} > ${ymk_sc_and_pvc_and_pod1}
    add_yaml_delimiter ${ymk_sc_and_pvc_and_pod1}
	cat ${yml_pvc} >> ${ymk_sc_and_pvc_and_pod1}
    add_yaml_delimiter ${ymk_sc_and_pvc_and_pod1}
    cat ${yml_pod1} >> ${ymk_sc_and_pvc_and_pod1}
    cat ${ymk_sc_and_pvc_and_pod1}

	echo "####### ---> ${S}. Run all in one yaml(SC, PVC and POD)"
    kubectl create -f ${ymk_sc_and_pvc_and_pod1}

	echo "## ---> ${S}.1. Verify PVC and PV info status and inpect"
    wait_for_item pvc $PVCName ${PVC_GOOD_STATUS} 5 2
    pvname=`kubectl get pvc $PVCName --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 5 2

 	echo "## ---> ${S}.2. Verify POD info status "
    wait_for_item pod $PODName Running 15 3

	echo "## ---> ${S}.3 Write DATA on the volume by create a file in /data inside the container"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"

	echo "## ---> ${S}.4 Delete all in one (SC, PVC, PV and POD)"
    kubectl delete -f ${ymk_sc_and_pvc_and_pod1}
    wait_for_item_to_delete pod $PODName 15 3
    wait_for_item_to_delete pvc $PVCName 10 2
    wait_for_item_to_delete pv $pvname 10 2
    wait_for_item_to_delete storageclass $profile 5 2
}

function basic_test_POD_with_2_volumes()
{
    # Description of the test :
    # -------------------------
    # This test start a POD with a container that consume 2 PVCs one on /data1 and other on /data2
    # The test validate that the container can see and use these 2 mountpoint.
    # -------------------------

	stepinc
	echo "########################### [Run 2 vols in the same POD-container] ###############"
	echo "####### ---> ${S}. Prepare yml with all the definition"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}

    yml_pvc1=$scripts/../deploy/scbe_volume_pvc_${PVCName}1.yml
    cp -f ${yml_pvc_template} ${yml_pvc1}
    sed -i -e "s/PVCNAME/${PVCName}1/g" -e "s/SIZE/1Gi/g" -e "s/PROFILE/$profile/g" ${yml_pvc1}
    yml_pvc2=$scripts/../deploy/scbe_volume_pvc_${PVCName}2.yml
    cp -f ${yml_pvc_template} ${yml_pvc2}
    sed -i -e "s/PVCNAME/${PVCName}2/g" -e "s/SIZE/1Gi/g" -e "s/PROFILE/$profile/g" ${yml_pvc2}

    yml_pod2=$scripts/../deploy/scbe_volume_with_pod2.yml
    cp -f ${yml_two_vols_pod_template} ${yml_pod2}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME1/${volPODname}1/g" -e "s|MOUNTPATH1|/data1|g" -e "s/PVCNAME1/${PVCName}1/g"  -e "s/VOLNAME2/${volPODname}2/g" -e "s|MOUNTPATH2|/data2|g" -e "s/PVCNAME2/${PVCName}2/g" ${yml_pod2}

	my_yml=$scripts/../deploy/scbe_volume_with_sc_2pvc_and_pod.yml
	cat ${yml_sc_profile} > ${my_yml}
    add_yaml_delimiter ${my_yml}
	cat ${yml_pvc1} >> ${my_yml}
    add_yaml_delimiter ${my_yml}
	cat ${yml_pvc2} >> ${my_yml}
    add_yaml_delimiter ${my_yml}
    cat ${yml_pod2} >> ${my_yml}
    cat ${my_yml}

	echo "####### ---> ${S}. Run all in one yaml(SC, 2 PVCs and POD with 2PVCs)"
    kubectl create -f ${my_yml}

	echo "## ---> ${S}.1. Verify PVC and PV info status and inpect"
    wait_for_item pvc ${PVCName}1 ${PVC_GOOD_STATUS} 5 2
    wait_for_item pvc ${PVCName}2 ${PVC_GOOD_STATUS} 5 2
    pvname1=`kubectl get pvc ${PVCName}1 --no-headers -o custom-columns=name:spec.volumeName`
    pvname2=`kubectl get pvc ${PVCName}2 --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv ${pvname1} ${PVC_GOOD_STATUS} 5 2
    wait_for_item pv ${pvname2} ${PVC_GOOD_STATUS} 5 2

 	echo "## ---> ${S}.2. Verify POD info status "
    wait_for_item pod $PODName Running 15 3

	echo "## ---> ${S}.3 Write DATA on the volume by create a file in /data inside the container"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data1"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data2"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data1/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data1/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data2/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data2/file_on_A9000_volume"


 	echo "## ---> ${S}.4. Verify 2 vols attached and mounted in the kubelet node"
    wwn1=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn ${pvname1}`
    wwn2=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn ${pvname2}`
	ssh root@$node1 "df | egrep ubiquity | grep $wwn1"
	ssh root@$node1 "df | egrep ubiquity | grep $wwn2"

	echo "## ---> ${S}.5 Delete all in one (SC, 2 PVCs, PV and POD)"
    kubectl delete -f ${my_yml}
    wait_for_item_to_delete pod $PODName 15 3
    wait_for_item_to_delete pvc ${PVCName}1 10 2
    wait_for_item_to_delete pvc ${PVCName}2 10 2
    wait_for_item_to_delete pv ${pvname1} 10 2
    wait_for_item_to_delete pv ${pvname2} 10 2
    wait_for_item_to_delete storageclass $profile 5 2
}

function fstype_basic_check()
{
    # TODO migrate to k8s style
    for fstype in ext4 xfs; do
        stepinc
        echo "####### ---> ${S}. Create volume with xfs run container and make sure the volume is $fstype"
        docker volume create --driver ubiquity --name $vol --opt size=5 --opt profile=${profile} --opt fstype=$fstype

        echo "## ---> ${S}.1. Verify volume info"
        docker volume ls | grep $vol
        docker volume inspect $vol | grep fstype | grep $fstype

        echo "## ---> ${S}.2 Run container with the volume and Verify it was mounted right"
        docker run -t -i -d --name ${CName}4 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine
        df | egrep "ubiquity|^Filesystem"
        mount |grep ubiquity | grep $fstype
        docker stop ${CName}4
        docker rm ${CName}4
        docker volume rm $vol
    done
}

function one_node_negative_tests()
{
    # TODO migrate to k8s style
	stepinc
	echo "####### ---> ${S}. some negative"
	echo "## ---> ${S}.1. Should fail to create volume with long name"
	long_vol_name=""; for i in `seq 1 63`; do long_vol_name="$long_vol_name${i}"; done
	docker volume create --driver ubiquity --name $long_vol_name --opt size=5 --opt profile=${profile} && exit 81 || :

	echo "## ---> ${S}.2. Should fail to create volume with wrong size"
	docker volume create --driver ubiquity --name $vol --opt size=10XX --opt profile=${profile} && exit 82 || :

	echo "## ---> ${S}.3. Should fail to create volume on wrong service"
	docker volume create --driver ubiquity --name $vol --opt size=10 --opt profile=${profile}XX && exit 83 || :
}


function tests_with_second_node()
{
    # TODO migrate to k8s style
	# Assuming plugin runs on second node and with storage connectivity
	echo ""
	echo "######### [2 nodes testing  node1=`hostname`, node2=`$node2`] ###########"

	stepinc
	echo "####### ---> ${S}. Run stateful container (should create and run the container)"
	docker run -t -i -d --name ${CName}4 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify volume was created for this container and you can touch a file inside the container"
	docker volume ls | grep $vol
	docker exec ${CName}4 touch /data/file_on_A9000_volume
	docker exec ${CName}4 ls -l /data/file_on_A9000_volume

	echo "## ---> ${S}.2. [$node2] : Verify volume is visible from second node"
	ssh root@$node2 "docker volume ls | grep $vol"

	echo "## ---> ${S}.3. [$node2] : Verify that you can NOT run container with $vol on second node"
	ssh root@$node2 "docker run -t -i -d --name ${CName}5 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine" && exit 1 || :
	ssh root@$node2 "docker stop ${CName}5"
	ssh root@$node2 "docker rm ${CName}5"

	echo "## ---> ${S}.4. [$node2] : Verify that you can NOT delete the volume $vol from the second node because its already attached to first node"
	ssh root@$node2 "docker volume rm $vol" && exit 1 || :
	ssh root@$node2 "docker volume ls | grep -v $vol" # volume should still be visible on the remote
	docker volume ls| grep -v $vol # and also visible on the local node, so we sure the volume was deleted

	stepinc
	echo "####### ---> ${S} Stop the container (so next step can run it on second node)"
	docker stop ${CName}4
	docker rm ${CName}4
	sleep 2 && echo "finished sleep 2 seconds"  # just waiting for detach to complite

	stepinc
	echo "####### ---> ${S} [$node2] : Start the container with the same vol on the second node"
	ssh root@$node2 "docker run -t -i -d --name ${CName}5 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine"

	echo "## ---> ${S}.1 [$node2] : Verify data presiste after migration to second node."
	ssh root@$node2 "docker exec ${CName}5 ls -l /data/file_on_A9000_volume"

	echo "## ---> ${S}.2 [$node2] : And add new file inside the volume."
	ssh root@$node2 "docker exec ${CName}5 touch /data/file_on_A9000_volume_from_node2"

	stepinc
	echo "####### ---> ${S}  [$node2] Stop the container on second node"
	ssh root@$node2 "docker stop ${CName}5"
	ssh root@$node2 "docker rm ${CName}5"

	stepinc
	echo "####### ---> ${S} [$node2] : Start the container with the same vol on the first node"
	docker run -t -i -d --name ${CName}6 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1 Verify data presiste after migration back to first node(check 2 files)."
	docker exec ${CName}6 ls -l /data/file_on_A9000_volume
	docker exec ${CName}6 ls -l /data/file_on_A9000_volume_from_node2

	stepinc
	echo "####### ---> ${S} Stop container and delete vol $vol"
	docker stop ${CName}6
	docker rm ${CName}6
	docker volume rm $vol

	echo "## ---> ${S}.1. [$node2] : Verify volume is no longer visible on the second node"
	ssh root@$node2 "docker volume ls | grep -v $vol "
}

function stepinc() { S=`expr $S + 1`; }

function setup()
{
    # TODO migrate to k8s style
    # clean acceptance containers and volumes before start the test and also validate ssh connection to second node if needed.
     conlist=`docker ps -a | grep $CName || :`
    if [ -n "$conlist" ]; then
       echo "Found $CName on the host `hostname`, try to stop and kill them before start the test"
       docker ps -a | grep $CName
       conlist2=`docker ps -a | sed '1d' | grep $CName | awk '{print $1}'|| :`
       docker stop $conlist2
       docker rm $conlist2
    fi

     volist=`docker volume ls -q | grep $CName || :`
    if [ -n "$volist" ]; then
       echo "Found $CName on the host, try to remove them"
       docker volume rm $volist
    fi

    if [ -n "$node2" ]; then
	ssh root@$node2 hostname || { echo "Cannot ssh to second host $node2, Aborting."; exit 1; }
        ssh root@$node2 "docker ps -aq | grep $CName" && { echo "need to clean $CName containers on remote node $node2"; exit 2; } || :
        ssh root@$node2 "docker volume ls | grep $CName" && { echo "need to clean $CName volumes on remote node $node2"; exit 3; } || :
    fi
}
[ "$1" = "-h" ] && { echo "$0 can get the following envs :"; echo "        ACCEPTANCE_PROFILE, ACCEPTANCE_WITH_NEGATIVE, ACCEPTANCE_WITH_SECOND_NODE"; exit 0; }
scripts=$(dirname $0)

S=0 # steps counter

[ -n "$ACCEPTANCE_PROFILE" ] && profile=$ACCEPTANCE_PROFILE || profile=gold
[ -n "$ACCEPTANCE_WITH_NEGATIVE" ] && withnegative=$ACCEPTANCE_WITH_NEGATIVE || withnegative=""
[ -n "$ACCEPTANCE_WITH_SECOND_NODE" ] && node2=$ACCEPTANCE_WITH_SECOND_NODE || node2=""
[ -n "$ACCEPTANCE_WITH_FIRST_NODE" ] && node1=$ACCEPTANCE_WITH_FIRST_NODE || { echo "env ACCEPTANCE_WITH_FIRST_NODE not provided. exit."; exit 1; }


yml_sc_tmplemt=$scripts/../deploy/scbe_volume_storage_class_template.yml
yml_pvc_template=$scripts/../deploy/scbe_volume_pvc_template.yml
yml_pod_template=$scripts/../deploy/scbe_volume_with_pod_template.yml
yml_two_vols_pod_template=$scripts/../deploy/scbe_volume_with_pod_with_2vols_template.yml

YAML_DELIMITER='---'
PVCName=accept-pvc
PODName=accept-pod
CName=accept-con # name of the containers in the script
vol=${CName}-vol
volPODname=accept-vol-pod
echo "Start Acceptance Test for IBM Block Storage"

setup # Verifications and clean up before the test

basic_tests_on_one_node
basic_tests_on_one_node_sc_pvc_pod_all_in_one
basic_test_POD_with_2_volumes

#fstype_basic_check
#[ -n "$withnegative" ] && one_node_negative_tests
#[ -n "$node2" ] && tests_with_second_node

echo ""
echo "======================================================"
echo "Successfully Finish The Acceptance test ([$S] steps). Running stateful container on IBM Block Storage."
