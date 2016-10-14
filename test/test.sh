#!/bin/bash

set -e

bin=$(dirname $0)

go build -o $bin/../out/ubiquity $bin/../cli.go

#create a fileset
printf "\n creating test volume \n "
mmcrfileset gpfs1 testvolume
cd $bin/../out
printf "\n calling spectrum attach \n "
./ubiquity attach \{\"volumeID\"\:\"testvolume\",\"filesystem\"\:\"gpfs1\",\"path\"\:\"/gpfs/gpfs1\"\}
printf "\n calling spectrum mount \n"
mkdir -p /tmp/dir1
./ubiquity mount /tmp/dir1/testvolume  testvolume \{\}
rm -rf /tmp/dir1
printf "\n calling spectrum unmount \n "
./ubiquity unmount /gpfs/gpfs1/testvolume
printf "\n calling spectrum detach \n "
./ubiquity detach testvolume
printf "\n cleaning test volume \n "
mmdelfileset gpfs1 testvolume -f

