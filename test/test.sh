#!/bin/bash

set -e

bin=$(dirname $0)

go build -o $bin/../out/spectrum $bin/../cli.go

#create a fileset
printf "\n creating test volume \n "
mmcrfileset gpfs1 testvolume
cd $bin/../out
printf "\n calling spectrum attach \n "
./spectrum attach \{\"volumeID\"\:\"testvolume\",\"filesystem\"\:\"gpfs1\",\"path\"\:\"/gpfs/gpfs1\"\}
printf "\n calling spectrum mount \n"
./spectrum mount /gpfs/gpfs1 testvolume \{\}
printf "\n calling spectrum unmount \n "
./spectrum unmount /gpfs/gpfs1/testvolume
printf "\n calling spectrum detach \n "
./spectrum detach testvolume
printf "\n cleaning test volume \n "
mmdelfileset gpfs1 testvolume

