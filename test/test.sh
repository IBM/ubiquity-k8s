#!/bin/bash

set -e

scripts=$(dirname $0)

cd $scripts/../bin
mmcrfileset gold testvolume

printf "\n calling ubiquity init \n"
./ubiquity init

printf "\n calling ubiquity mount \n"
./ubiquity attach \{\"volumeID\"\:\"testvolume\",\"filesystem\"\:\"gold\",\"path\"\:\"/gpfs/gold\"\}
printf "\n calling ubiquity mount \n"
mkdir -p /tmp/dir1
./ubiquity mount /tmp/dir1/testvolume  testvolume \{\}
rm -rf /tmp/dir1
printf "\n calling ubiquity unmount \n "
./ubiquity unmount /gpfs/gold/testvolume
printf "\n calling ubiquity detach \n "
./ubiquity detach testvolume
mmdelfileset gold testvolume
