#!/bin/bash -ex

############################################
# Acceptance Test for IBM Block Storage via SCBE
#
# Test cases:
#  ==========
#    One minion tests:
#    1. Create SC and PVC then create POD with volume, write data, stop POD and start again. Validate every step in the way.
#    2. Create SC, PVC, POD all in one yml file and delete all together. Validate every step in the way.
#    3. Create POD with 2 volumes. Validate every step in the way.
#    4. Create POD for each FSTYPE supported and delete them all. Validate fstype are correct.
#    Two minion tests:
#    5. Create POD with PVC1 on node1 then delete it and start it on node2. Validate PVC1 is no attached to node2.
#
# Script prerequisites:
#  ====================
#    1. SCBE server up and running with 1 service delegated to ubiquity interface (service name given by $ACCEPTANCE_PROFILE)
#    2. ubiqutiy server up and running with SCBE backend configured
#    3. ubiquity-provisioner up and running and also ubiquity-flexvolume-cli locate on the k8s nodes
#    4. setup connectivity between the minions to the related storage system of the service.
#    5. root SSH passwordless between master and minions (so test will validate on the minions the mount devices via ssh)
#    6. You must spseicy the menion that the PODs will run on by using env: export ACCEPTANCE_WITH_FIRST_NODE=nodename1
#    7. To enable migration tests between 2 nodes, you must specify the second minion by using env: export ACCEPTANCE_WITH_SECOND_NODE=nodename2
#    8. This script uses template yml files from ../deploy directory. So make sure you have this directory.
#
#
#  How to run the tests:
#  =====================
#     $> git clone https://github.com/IBM/ubiquity-k8s.git
#     $> cd ubiquity-k8s.git/scripts
#     [root@k8s-master ~]# kubectl get nodes
#     NAME         STATUS         AGE
#     k8s-master   Ready,master   15d
#     k8s-node1    Ready          15d
#     k8s-node2    Ready          1d
#
#     $> export ACCEPTANCE_WITH_FIRST_NODE=k8s-node1
#     $> export ACCEPTANCE_WITH_SECOND_NODE=k8s-node2
#
#     $> ./scripts/run_acceptance_ibm_block_storage_via_scbe.sh
#
#     Note : You should see "Successfully Finish The Acceptance test" if all tests passed OK
############################################


function basic_tests_on_one_node()
{
    # Description of the test :
    # -------------------------
    # The test creates and validate the following objects: SC, PVC, PV, POD with PVC.
    # The test validate that the relevant mount, multipathing was created after POD is up.
    # Then generate some IO on the volume inside the container, delete the POD and start it again
    # and validate that the data still persist.
    # -------------------------

	stepinc
	printf "\n\n\n\n"
	echo "########################### [basic POD test with PVC] ###############"
	echo "####### ---> ${S}. Creating Storage class for ${profile} service"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/SCNAME/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}
    cat $yml_sc_profile
    # kubectl create -f ${yml_sc_profile} || true
    if ! kubectl get storageclass $profile >/dev/null 2>&1; then
        kubectl create -f ${yml_sc_profile}
    else
        echo "Storage class $profile already exist, so no need to create it."
    fi
    kubectl get storageclass $profile

	echo "####### ---> ${S}. Create PVC (volume) on SCBE ${profile} service (which is on IBM FlashSystem A9000R)"
    yml_pvc=$scripts/../deploy/scbe_volume_pvc_${PVCName}.yml
    cp -f ${yml_pvc_template} ${yml_pvc}
    sed -i -e "s/PVCNAME/$PVCName/g" -e "s/SIZE/5Gi/g" -e "s/SCNAME/$profile/g" ${yml_pvc}
    cat ${yml_pvc}
    kubectl create -f ${yml_pvc}

	echo "####### ---> ${S}.1. Verify PVC and PV info status and inpect"
    wait_for_item pvc $PVCName ${PVC_GOOD_STATUS} 10 3

    pvname=`kubectl get pvc $PVCName --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 10 3
    kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname

    wwn=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname`
    kubectl get pv -o json $pvname | grep -A15 flexVolume
    kubectl get pv --no-headers -o custom-columns=column:metadata.annotations.Provisioner_Id $pvname | grep ubiquity-k8s-provisioner

	echo "## ---> ${S}.2. Verify storage side : verify the volume was created on the relevant pool\service"
	echo "Skip step"
	## ssh root@gen4d-67a "xcli.py vol_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run POD [$PODName] with container ${CName} with the new volume"
    yml_pod1=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod1}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" -e "s/NODESELECTOR/${node1}/g" ${yml_pod1}
    cat $yml_pod1
    kubectl create -f ${yml_pod1}
    wait_for_item pod $PODName Running 1120 3


	echo "## ---> ${S}.1. Verify the volume was attached to the kubelet node $node1"
	ssh root@$node1 "df | egrep ubiquity | grep '$wwn'"
	ssh root@$node1 "multipath -ll | grep -i $wwn"
	ssh root@$node1 'lsblk | egrep "ubiquity|^NAME" -B 1 | grep "$wwn"'
	ssh root@$node1 'mount |grep "$wwn"| grep "$fstype"'

	echo "## ---> ${S}.2. Verify volume exist inside the container"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data"

	echo "## ---> ${S}.3. Verify container with the mount point"
    kubectl describe pod $PODName | grep -A1 "Mounts"

	echo "## ---> ${S}.3. Verify the storage side : check volume has mapping to the host"
    echo "Skip step"
    ## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Write DATA on the volume by create a file in /data inside the container"
        # Add : at the end of touch, because kubectl has a known exit code issue.
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data/file_on_A9000_volume" || :
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"

	stepinc
	echo "####### ---> ${S}. Stop the container"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 1120 3

	echo "## ---> ${S}.1. Verify the volume was detached from the kubelet node"
	sleep 2 # some times mount is not refreshed immediate
	ssh root@$node1 "df | egrep ubiquity | grep $wwn" && exit 1 || :
	ssh root@$node1 "multipath -ll | grep -i $wwn" && exit 1 || :
	ssh root@$node1 "lsblk | egrep ubiquity -B 1 | grep $wwn" && exit 1 || :
	ssh root@$node1 "mount |grep '$wwn' | grep '$fstype'" && exit 1 || :

	echo "## ---> ${S}.2. Verify PVC and PV still exist"
    kubectl get pvc $PVCName
    kubectl get pv $pvname

	echo "## ---> ${S}.3. Verify the storage side : check volume is no longer mapped to the hos"
    echo "Skip step"
	## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run the POD again with the same volume and check the if the data remains"
    kubectl create -f ${yml_pod1}
    wait_for_item pod $PODName Running 1120 3

	echo "## ---> ${S}.1. Verify that the data remains (file exist on the /data inside the container)"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"


	stepinc
	echo "####### ---> ${S}. Stop the POD"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 1120 3

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
    wait_for_item_to_delete storageclass $profile 10 3
}

function basic_tests_on_one_node_sc_pvc_pod_all_in_one()
{
    # Description of the test :
    # -------------------------
    # This test generates yml file with definition of many object type, such as SC, PVC and POD
    # Then create these object in one kubectl command and validate all is up and running.
    # -------------------------

	stepinc
	printf "\n\n\n\n"
	echo "########################### [All in one suite] ###############"
	echo "####### ---> ${S}. Prepare all in one yaml with SC, PVC, POD yml"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/SCNAME/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}

    yml_pvc=$scripts/../deploy/scbe_volume_pvc_${PVCName}.yml
    cp -f ${yml_pvc_template} ${yml_pvc}
    sed -i -e "s/PVCNAME/$PVCName/g" -e "s/SIZE/5Gi/g" -e "s/SCNAME/$profile/g" ${yml_pvc}

    yml_pod1=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod1}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" -e "s/NODESELECTOR/${node1}/g" ${yml_pod1}

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
    wait_for_item pvc $PVCName ${PVC_GOOD_STATUS} 10 3
    pvname=`kubectl get pvc $PVCName --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 10 3

 	echo "## ---> ${S}.2. Verify POD info status "
    wait_for_item pod $PODName Running 1120 3

	echo "## ---> ${S}.3 Write DATA on the volume by create a file in /data inside the container"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data/file_on_A9000_volume" || :
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data/file_on_A9000_volume"

	echo "## ---> ${S}.4 Delete all in one (SC, PVC, PV and POD)"
    kubectl delete -f ${ymk_sc_and_pvc_and_pod1}
    wait_for_item_to_delete pod $PODName 1120 3
    wait_for_item_to_delete pvc $PVCName 10 2
    wait_for_item_to_delete pv $pvname 10 2
    wait_for_item_to_delete storageclass $profile 10 3
}

function basic_test_POD_with_2_volumes()
{
    # Description of the test :
    # -------------------------
    # This test start a POD with a container that consume 2 PVCs one on /data1 and other on /data2
    # The test validate that the container can see and use these 2 mountpoint.
    # -------------------------

	stepinc
	printf "\n\n\n\n"
	echo "########################### [Run 2 vols in the same POD-container] ###############"
	echo "####### ---> ${S}. Prepare yml with all the definition"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/SCNAME/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}

    yml_pvc1=$scripts/../deploy/scbe_volume_pvc_${PVCName}1.yml
    cp -f ${yml_pvc_template} ${yml_pvc1}
    sed -i -e "s/PVCNAME/${PVCName}1/g" -e "s/SIZE/1Gi/g" -e "s/SCNAME/$profile/g" ${yml_pvc1}
    yml_pvc2=$scripts/../deploy/scbe_volume_pvc_${PVCName}2.yml
    cp -f ${yml_pvc_template} ${yml_pvc2}
    sed -i -e "s/PVCNAME/${PVCName}2/g" -e "s/SIZE/1Gi/g" -e "s/SCNAME/$profile/g" ${yml_pvc2}

    yml_pod2=$scripts/../deploy/scbe_volume_with_pod2.yml
    cp -f ${yml_two_vols_pod_template} ${yml_pod2}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME1/${volPODname}1/g" -e "s|MOUNTPATH1|/data1|g" -e "s/PVCNAME1/${PVCName}1/g"  -e "s/VOLNAME2/${volPODname}2/g" -e "s|MOUNTPATH2|/data2|g" -e "s/PVCNAME2/${PVCName}2/g" -e "s/NODESELECTOR/${node1}/g" ${yml_pod2}

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
    wait_for_item pvc ${PVCName}1 ${PVC_GOOD_STATUS} 10 3
    wait_for_item pvc ${PVCName}2 ${PVC_GOOD_STATUS} 10 3
    pvname1=`kubectl get pvc ${PVCName}1 --no-headers -o custom-columns=name:spec.volumeName`
    pvname2=`kubectl get pvc ${PVCName}2 --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv ${pvname1} ${PVC_GOOD_STATUS} 10 3
    wait_for_item pv ${pvname2} ${PVC_GOOD_STATUS} 10 3

 	echo "## ---> ${S}.2. Verify POD info status "
    wait_for_item pod $PODName Running 1120 3

	echo "## ---> ${S}.3 Write DATA on the volume by create a file in /data inside the container"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data1"
    kubectl exec -it  $PODName -c ${CName} -- bash -c "df /data2"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data1/file_on_A9000_volume" || :
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data1/file_on_A9000_volume"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch /data2/file_on_A9000_volume" || :
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l /data2/file_on_A9000_volume"


 	echo "## ---> ${S}.4. Verify 2 vols attached and mounted in the kubelet node"
    wwn1=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn ${pvname1}`
    wwn2=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn ${pvname2}`
	ssh root@$node1 "df | egrep ubiquity | grep '$wwn1'"
	ssh root@$node1 "df | egrep ubiquity | grep '$wwn2'"

	echo "## ---> ${S}.5 Delete all in one (SC, 2 PVCs, PV and POD)"
    kubectl delete -f ${my_yml}
    wait_for_item_to_delete pod $PODName 1120 3
    wait_for_item_to_delete pvc ${PVCName}1 10 2
    wait_for_item_to_delete pvc ${PVCName}2 10 2
    wait_for_item_to_delete pv ${pvname1} 10 2
    wait_for_item_to_delete pv ${pvname2} 10 2
    wait_for_item_to_delete storageclass $profile 10 3
}

function fstype_basic_check()
{

    # Description of the test :
    # -------------------------
    # This test start create 2 SCs(goldext4 goldxfs) and then start POD with one container
    # that uses 2 PVCs, one from each SCs. The test validate fstype of each PV mount.
    # -------------------------

	stepinc
    printf "\n\n\n\n"
	echo "########################### [ Run Pod with volume per fstype [${FS_SUPPORTED}] ] ###############"
	echo "####### ---> ${S}. Prepare yml with all the definition"

  	my_yml=$scripts/../deploy/scbe_volume_with_2sc_2pvc_and_pod.yml
    cat /dev/null > ${my_yml}
    for fstype in ${FS_SUPPORTED}; do
        echo "Generate yml for SC and PVC and POD according to the $fstype"
        yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_${profile}_${fstype}.yml
        cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
        sed -i -e "s/PROFILE/$profile/g" -e "s/SCNAME/${profile}-${fstype}/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}
        add_yaml_delimiter ${yml_sc_profile}
	    cat ${yml_sc_profile} >> ${my_yml}

        yml_pvc1=$scripts/../deploy/scbe_volume_pvc_${PVCName}_${fstype}.yml
        cp -f ${yml_pvc_template} ${yml_pvc1}
        sed -i -e "s/PVCNAME/${PVCName}-${fstype}/g" -e "s/SIZE/1Gi/g" -e "s/SCNAME/${profile}-${fstype}/g" ${yml_pvc1}
        add_yaml_delimiter ${yml_pvc1}
	    cat ${yml_pvc1} >> ${my_yml}

        yml_pod1=$scripts/../deploy/scbe_volume_with_pod_with_pvc_for_each_fstype.yml
        cp -f ${yml_pod_template} ${yml_pod1}
        sed -i -e "s/PODNAME/${PODName}-${fstype}/g" -e "s/CONNAME/${CName}-${fstype}/g"  -e "s/VOLNAME/${volPODname}-${fstype}/g" -e "s|MOUNTPATH|/data-${fstype}|g" -e "s/PVCNAME/$PVCName-${fstype}/g" -e "s/NODESELECTOR/${node1}/g" ${yml_pod1}
        add_yaml_delimiter ${yml_pod1}
	    cat ${yml_pod1} >> ${my_yml}
    done
    echo "Here is the final yml file with all SC, PVC, POD for ${FS_SUPPORTED}:"
    cat ${my_yml}

  	stepinc
    echo "####### ---> ${S}. Run all in one yaml(SC, PVCs and POD for each fstype)"
    kubectl create -f ${my_yml}


    for fstype in ${FS_SUPPORTED}; do
        echo "## ---> ${S}.1. Verify PVC, PV and POD of $fstype"
        wait_for_item pvc ${PVCName}-${fstype} ${PVC_GOOD_STATUS} 10 3
        pvname1=`kubectl get pvc ${PVCName}-${fstype} --no-headers -o custom-columns=name:spec.volumeName`
        wait_for_item pv ${pvname1} ${PVC_GOOD_STATUS} 10 3
        wait_for_item pod ${PODName}-${fstype} Running 1120 3


        echo "## ---> ${S}.2. Verify POD $fstype really mounted $fstype filesystem"
        # Should fail if its not the right fstype
        kubectl exec -it ${PODName}-${fstype} -c ${CName}-${fstype} -- bash -c "mount | grep data-${fstype}" | awk '{print $5}' | grep ${fstype}
    done

  	stepinc
	echo "## ---> ${S} Delete all in one (SC, PVCs, PV and POD for each fstype : ${FS_SUPPORTED})"
    kubectl delete -f ${my_yml}
    for fstype in ${FS_SUPPORTED}; do
        wait_for_item_to_delete pod ${PODName}-${fstype} 1120 3
        wait_for_item_to_delete pvc ${PVCName}-${fstype} 10 2
        wait_for_item_to_delete storageclass ${profile}-${fstype} 10 3
    done
    
    # Now also wait for PVs is still exist
    # pvs=`kubectl get pv --no-headers -o custom-columns=wwn:metadata.name`
    # [ -z "$pvs" ] && return
    # for pv in $pvs; do
    #     wait_for_item_to_delete pv $pv 10 2
    # done
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
    # Description of the test :
    # -------------------------
    # The test creates and validate the following objects: SC, PVC, PV, POD with PVC(on node1)
    # the delete the POD and run the same one with the same PVC on node2. Then validate the migration.
    # -------------------------

	printf "\n\n\n\n"
	echo "########################### [Migrate POD from node1 to node2 ] ###############"
    [ -z "$node2" ] && { echo "Skip running migration test - because env ACCEPTANCE_WITH_SECOND_NODE was not set."; return; }

	stepinc
	echo "####### ---> ${S} Steps on NODE1 $node1"

	echo "## ---> ${S}.1 Creating Storage class for ${profile} service"
    yml_sc_profile=$scripts/../deploy/scbe_volume_storage_class_$profile.yml
    cp -f ${yml_sc_tmplemt} ${yml_sc_profile}
    fstype=ext4
    sed -i -e "s/PROFILE/$profile/g" -e "s/SCNAME/$profile/g" -e "s/FSTYPE/$fstype/g" ${yml_sc_profile}
    cat $yml_sc_profile
    kubectl create -f ${yml_sc_profile}
    kubectl get storageclass $profile

	echo "## --> ${S}.2 Create PVC (volume) on SCBE ${profile} service (which is on IBM FlashSystem A9000R)"
    yml_pvc=$scripts/../deploy/scbe_volume_pvc_${PVCName}.yml
    cp -f ${yml_pvc_template} ${yml_pvc}
    sed -i -e "s/PVCNAME/$PVCName/g" -e "s/SIZE/5Gi/g" -e "s/SCNAME/$profile/g" ${yml_pvc}
    cat ${yml_pvc}
    kubectl create -f ${yml_pvc}

	echo "## ---> ${S}.3. Verify PVC and PV info status and inpect"
    wait_for_item pvc $PVCName ${PVC_GOOD_STATUS} 10 3
    pvname=`kubectl get pvc $PVCName --no-headers -o custom-columns=name:spec.volumeName`
    wait_for_item pv $pvname ${PVC_GOOD_STATUS} 10 3
    kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname
    wwn=`kubectl get pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $pvname`

	echo "## ---> ${S}.4. Run POD [$PODName] with container ${CName} with the new volume"
    yml_pod1=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod1}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" -e "s/NODESELECTOR/${node1}/g" ${yml_pod1}
    cat $yml_pod1
    kubectl create -f ${yml_pod1}
    wait_for_item pod $PODName Running 120 3


	echo "## ---> ${S}.5. Verify the volume was attached to the kubelet node $node1"
	ssh root@$node1 "df | egrep ubiquity | grep '$wwn'"
	ssh root@$node1 "multipath -ll | grep -i $wwn"
	ssh root@$node1 'lsblk | egrep "ubiquity|^NAME" -B 1'
	ssh root@$node1 "mount |grep '$wwn' | grep '$fstype'"

	echo "## ---> ${S}.6 Write DATA on the volume by create a file in /data inside the container"
        file_create_node1="/data/file_created_on_${node1}"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "touch ${file_create_node1}" || :
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l ${file_create_node1}"

	stepinc
	echo "####### ---> ${S}. Stop the container on $node1"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 1120 3

	echo "## ---> ${S}.1. Verify the volume was detached from the kubelet node"
	sleep 2 # some times mount is not refreshed immediate
	ssh root@$node1 "df | egrep ubiquity | grep $wwn" && exit 1 || :

	echo "## ---> ${S}.2. Verify PVC and PV still exist"
       kubectl get pvc $PVCName
       kubectl get pv $pvname


	stepinc
	echo "####### ---> ${S} Steps on NODE2 $node2"

	echo "## ---> ${S}.1 Run the POD again BUT now on second node : $node2"
    yml_pod2=$scripts/../deploy/scbe_volume_with_pod1.yml
    cp -f ${yml_pod_template} ${yml_pod2}
    sed -i -e "s/PODNAME/$PODName/g" -e "s/CONNAME/$CName/g"  -e "s/VOLNAME/$volPODname/g" -e "s|MOUNTPATH|/data|g" -e "s/PVCNAME/$PVCName/g" -e "s/NODESELECTOR/${node2}/g" ${yml_pod2}
    cat $yml_pod2
    kubectl create -f ${yml_pod2}
    wait_for_item pod $PODName Running 120 3

	echo "## ---> ${S}.2. Verify that the data remains (file exist on the /data inside the container)"
	kubectl exec -it  $PODName -c ${CName} -- bash -c "ls -l ${file_create_node1}"


	stepinc
	echo "####### ---> ${S}. Stop the POD"
    kubectl delete -f ${yml_pod1}
    wait_for_item_to_delete pod $PODName 1120 3

	stepinc
	echo "####### ---> ${S}. Remove the PVC and PV"
	kubectl delete -f ${yml_pvc}
    wait_for_item_to_delete pvc $PVCName 10 2
    wait_for_item_to_delete pv $pvname 10 2

	stepinc
	echo "####### ---> ${S}. Remove the Storage Class $profile"
    kubectl delete -f ${yml_sc_profile}
    wait_for_item_to_delete storageclass $profile 10 3
}

function setup()
{
	echo "####### ---> ${S}. Verify that no volume attached to the kube node1"
    wwn=`kubectl get $nsf pv --no-headers -o custom-columns=wwn:spec.flexVolume.options.Wwn $POSTGRES_PV`
	ssh root@$node1 'df | egrep "ubiquity" | grep -v $wwn' && exit 1 || :
	ssh root@$node1 'multipath -ll | grep IBM | grep -v $wwn' && exit 1 || :
	ssh root@$node1 'lsblk | egrep "ubiquity" -B 1 | grep -v $wwn' && exit 1 || :
	kubectl get $nsf pvc 2>&1 | grep "$POSTGRES_PV"
	kubectl get $nsf pv 2>&1 | grep "$POSTGRES_PV"

    echo "Skip clean up the environment for acceptance test (TODO)"
    return

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
function usage()
{
    echo "Usage $> $0 [ubiquity-namespace] [-h]"
    echo "    [ubiquity-namespace] : The namespace where Ubiqutiy is running."
    echo "    -h : Print this usage."
    echo "    Environment variables:";
    echo "        export ACCEPTANCE_PROFILE=<The SCBE profile name to work with>"
    echo "        export ACCEPTANCE_WITH_NEGATIVE=<bool>. true in order to run additional negative tests"
    echo "        export ACCEPTANCE_WITH_FIRST_NODE=<IP of first minion>"
    echo "        export ACCEPTANCE_WITH_SECOND_NODE=<IP of second minion for volume migration scenario>"
    # TODO : should refactor usage with flags -n <namespace> -p <scbe profile> -f <first node IP> -s <second node IP> -g (for negative)
    exit 1
}

[ "$1" = "-h" ] && { usage; }
[ -n "$1" ] && NS=$1 || NS=ubiquity
nsf="--namespace $NS"
echo "Assume Ubiquity namespace is [$NS]"
scripts=$(dirname $0)

S=0 # steps counter

. $scripts/acceptance_utils.sh

[ -n "$ACCEPTANCE_PROFILE" ] && profile=$ACCEPTANCE_PROFILE || profile=gold
[ -n "$ACCEPTANCE_WITH_NEGATIVE" ] && withnegative=$ACCEPTANCE_WITH_NEGATIVE || withnegative=""
[ -n "$ACCEPTANCE_WITH_SECOND_NODE" ] && node2=$ACCEPTANCE_WITH_SECOND_NODE || node2=""
[ -n "$ACCEPTANCE_WITH_FIRST_NODE" ] && node1=$ACCEPTANCE_WITH_FIRST_NODE || { echo "env ACCEPTANCE_WITH_FIRST_NODE not provided. exit."; exit 1; }


yml_sc_tmplemt=$scripts/../deploy/scbe_volume_storage_class_template.yml
yml_pvc_template=$scripts/../deploy/scbe_volume_pvc_template.yml
yml_pod_template=$scripts/../deploy/scbe_volume_with_pod_template.yml
yml_two_vols_pod_template=$scripts/../deploy/scbe_volume_with_pod_with_2vols_template.yml

POSTGRES_PV="ibm-ubiquity-db"
FS_SUPPORTED="ext4 xfs"
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
fstype_basic_check
#[ -n "$withnegative" ] && one_node_negative_tests
tests_with_second_node

echo ""
echo "======================================================"
echo "Successfully Finish The Acceptance test ([$S] steps). Running stateful container on IBM Block Storage."
