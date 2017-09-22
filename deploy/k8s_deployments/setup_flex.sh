#!/bin/sh

rm -rf /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
rm /etc/ubiquity/ubiquity-k8s-flex.conf

cp   ubiquity-k8s-flex /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/
cp   /var/tmp/ubiquity-config/ubiquity-k8s-flex.conf /etc/ubiquity/
while :
do
	sleep 1000
done