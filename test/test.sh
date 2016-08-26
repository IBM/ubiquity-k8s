#!/bin/bash

set -e

bin=$(dirname $0)

go build -o $bin/../out/spectrum $bin/../cli.go

#create a fileset
echo "creating test volume"
mmcrfileset gpfs1 testvolume
cd $bin/../out
echo "calling spectrum attach"
./spectrum attach \{\"volumeID\"\:\"testvolume\",\"filesystem\"\:\"gpfs1\"\,\"path\"\:\"/gpfs/gpfs1\"\}
echo "calling spectrum mount"
./spectrum mount /gpfs/gpfs1 testvolume
echo "calling spectrum unmount"
./spectrum unmount /gpfs/gpfs1/testvolume
echo "calling spectrum detach"
./spectrum detach testvolume
echo "cleaning test volume"
mmdelfileset gpfs1 testvolume

