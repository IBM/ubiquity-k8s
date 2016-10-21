#!/bin/bash

set -e

bin=$(dirname $0)

go build -o $bin/../out/ubiquity $bin/../cli.go

cd $bin/../out
mmcrfileset gold testvolume
# printf "\n calling spectrum attach \n "
./ubiquity attach \{\"volumeID\"\:\"testvolume\",\"filesystem\"\:\"gold\",\"path\"\:\"/gpfs/gold\"\}
printf "\n calling spectrum mount \n"
mkdir -p /tmp/dir1
./ubiquity mount /tmp/dir1/testvolume  testvolume \{\}
rm -rf /tmp/dir1
printf "\n calling spectrum unmount \n "
./ubiquity unmount /gpfs/gold/testvolume
printf "\n calling spectrum detach \n "
./ubiquity detach testvolume
mmdelfileset gold testvolume
