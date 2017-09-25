#!/bin/sh

echo "INFO: ubiquity-flex: removing previous versions of flex plugin and configuration"
rm -rf /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/ubiquity-k8s-flex
rm /etc/ubiquity/ubiquity-k8s-flex.conf

echo "INFO: ubiquity-flex: copying new versions of flex plugin and configuration"
cp   ubiquity-k8s-flex /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/
cp   /var/tmp/ubiquity-config/ubiquity-k8s-flex.conf /etc/ubiquity/

echo "INFO: ubiquity-flex: sleeping"
while :
do
    echo "INFO: ubiquity-flex: Pod sleeping 1000, so far so good"
	sleep 1000
done