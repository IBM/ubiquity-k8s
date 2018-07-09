#!/bin/bash

set -e

scripts=$(dirname $0)

cd $scripts/../bin
sudo mmcrfileset gold testvolume

printf "\n calling ubiquity init \n"
./ubiquity-k8s-flex init

printf "\n calling ubiquity mount \n"
./ubiquity-k8s-flex attach \{\"volumeName\"\:\"testvolume\",\"fileset\"\:\"testvolume\",\"filesystem\"\:\"gold\",\"path\"\:\"/gpfs/gold\"\}
printf "\n calling ubiquity mount \n"
mkdir -p /tmp/dir1
./ubiquity-k8s-flex mount /tmp/dir1/testvolume  testvolume \{\}
rm -rf /tmp/dir1
printf "\n calling ubiquity unmount \n "
./ubiquity-k8s-flex unmount /gpfs/gold/testvolume
printf "\n calling ubiquity detach \n "
./ubiquity-k8s-flex detach testvolume
sudo mmdelfileset gold testvolume
