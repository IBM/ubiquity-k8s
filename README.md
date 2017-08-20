# Ubiquity Kubernetes Persistent Storage
This project includes components for managing [Kubernetes persistent storage](https://kubernetes.io/docs/concepts/storage/persistent-volumes), using [Ubiquity](https://github.com/IBM/ubiquity) service.
- [Ubiquity Dynamic Provisioner](ubiquity-dynamic-provisioner) for creating and deleting persistent volumes
- [Ubiquity FlexVolume Driver CLI](ubiquity-flexvolume-cli) for attaching and detaching persistent volumes

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Ubiquity Dynamic Provisioner 

Ubiquity Dynamic Provisioner (Provisioner) is intended for creation and deletion of persistent volumes in Kubernetes, using the  [Ubiquity](https://github.com/IBM/ubiquity) service.
  
### Installing the Ubiquity Dynamic Provisioner
Install and configure the Provisioner on a single node in the Kubernetes cluster (minion or master).

### 1. Prerequisites
  * The Provisioner is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

  * The Provisioner requires Kubernetes version 1.5.6 or above.

  * The following sudoers configuration `/etc/sudoers` is required to run the Provisioner as root user: 
  
     ```
        Defaults !requiretty
     ```
     For non-root users, such as USER, configure the sudoers as follows: 

     ```
         USER ALL= NOPASSWD: /usr/bin/, /bin/, /usr/sbin/ 
         Defaults:%USER !requiretty
         Defaults:%USER secure_path = /sbin:/bin:/usr/sbin:/usr/bin
     ```
  * The Provisioner requires the Kubernetes config file to be available at `~/.kube/config`. For example, to generate the file:
     ```bash
         mkdir ~/.kube
         cp /etc/kubernetes/admin.conf ~/.kube/config
     ```
  * Opening TCP ports to Ubiquity server:

    Ubiquity server listens on TCP port (by default 9999) to receive the Provisioner requests, such as creating a new volume. Verify that the node can access this Ubiquity server port.


### 2. Downloading and installing the Provisioner

* Download and unpack the application package.
     ```bash
         mkdir -p /etc/ubiquity
         cd /etc/ubiquity
         curl -L https://github.com/IBM/ubiquity-k8s/releases/download/v0.4.0/ubiquity-k8s-provisioner-0.4.0.tar.gz | tar xf -
         cp ubiquity-k8s-provisioner /usr/bin
         chmod u+x /usr/bin/ubiquity-k8s-provisioner
         #chown USER:GROUP /usr/bin/ubiquity-k8s-provisioner   ### Run this command only a non-root user.
         cp ubiquity-k8s-provisioner.service /usr/lib/systemd/system/ 
     ```
* To run the Provisioner as non-root user, add the `User=USER` line under the [Service] item in the  `/usr/lib/systemd/system/ubiquity-k8s-provisioner.service` file.
   
* Enable the Provisioner service.
    ```bash
        systemctl enable ubiquity-k8s-provisioner.service
    ```

### 3. Configuring the Provisioner
Before running the Provisioner service, create and configure the `/etc/ubiquity/ubiquity-k8s-provisioner.conf` file.

Configuration file example. (Make sure to set the `backend` with `scbe` for IBM block storage usage or `spectrum-scale` for IBM Spectrum Scale usage.)
```toml
logPath = "/var/tmp"  # The Ubiquity Provisioner writes logs to the "ubiquity-k8s-provisioner.log" file.
backends = ["scbe"] # Backend name, such as scbe or spectrum-scale.
logLevel = "info" # Optional parameter. Possible values are debug, info or error. Default is "info".


[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity service.
port = 9999            # TCP port on which the Ubiquity service is listening.
```
  * Verify that the logPath, exists on the host before starting the Provisioner.


### 5. Running the Provisioner service
  * Run the service.
    ```bash
        systemctl start ubiquity-k8s-provisioner

        # Verify that the service is `active (running)`
        systemctl status ubiquity-k8s-provisioner
    ```

<br>
<br>
<br>
<br>

## Ubiquity FlexVolume Driver CLI 

Ubiquity FlexVolume Driver CLI (FlexVolume) supports attaching and detaching volumes on persistent storage, using the [Ubiquity](https://github.com/IBM/ubiquity) service.

### Installing the FlexVolume
Install and configure the FlexVolume on each minion node in the Kubernetes cluster that requires access to Ubiquity volumes.

### 1. Prerequisites
  * The FlexVolume is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

  * The FlexVolume requires Kubernetes version 1.5.6 or above.

  * Kubernetes version 1.6 introduced remote volume attachment and detachment from the Kubernetes controller. This functionality is not yet supported by Ubiquity. To avoid any issues, set `--enable-controller-attach-detach` to `false`. Perform this procedure in the `/etc/systemd/system/kubelet.service.d/10-kubeadm.conf` file as follows:
    
    ```bash
     Environment="KUBELET_KUBECONFIG_ARGS=--kubeconfig=/etc/kubernetes/kubelet.conf --require-kubeconfig=true --enable-controller-attach-detach=false"
    ```

  * The following sudoers configuration `/etc/sudoers` is required to run the FlexVolume as root user: 
  
     ```
        Defaults !requiretty
     ```
     For non-root users, such as USER, configure the sudoers as follows: 

     ```
         USER ALL= NOPASSWD: /usr/bin/, /bin/, /usr/sbin/ 
         Defaults:%USER !requiretty
         Defaults:%USER secure_path = /sbin:/bin:/usr/sbin:/usr/bin
     ```

  * The Kubernetes node must have access to the storage backends. Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.

  * Opening TCP ports to Ubiquity server:

    Ubiquity server listens on TCP port (by default 9999) to receive the FlexVolume requests, such as volume attachment. Verify that the node can access this Ubiquity server port.


### 2. Downloading and installing the FlexVolume

* Download and unpack the application package.
     ```bash
         mkdir -p /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
         cd /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex
         curl -L https://github.com/IBM/ubiquity-k8s/releases/download/v0.4.0/ubiquity-k8s-flex
         chmod u+x ubiquity-k8s-flex
         #chown USER:GROUP ubiquity-k8s-flex   ### Run this command only for non-root user.
     ```

### 3. Configuring the FlexVolume
* Before using the FlexVolume, create and configure the `/etc/ubiquity/ubiquity-k8s-flex.conf` file (create the `/etc/ubiquity` directory if it does not exist), according to your storage system type.
Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section.

* Generic configuration file example:
```toml
logPath = "/var/tmp"  # The Ubiquity FlexVolume will write logs to file "ubiquity-k8s-flex.log" in this path.
backends = ["scbe"] # Backend name, such as scbe or spectrum-scale.
logLevel = "info" # Optional parameter. Possible values are debug, info or error. Default is "info".

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity service
port = 9999            # TCP port on which the Ubiquity service is listening
```


    * Verify that the logPath, exists on the host so the FlexVolume will be able to run properly.

### 4. Restart the kubelet to reload the new FlexVolume. 
* To reload the new FlexVolume that was located in kubelet-plugins directory, restart the `kubelete` service:
     ```bash
         systemctl restart kubelet
     ```

* Verify that the FlexVolume is functional.
     ```bash
         /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/ubiquity-k8s-flex init
         {"status":"Success","message":"Plugin init successfully","device":"","volumeName":"","attached":false}
     ```


<br>
<br>
<br>
<br>

## Usage examples for Ubiquity Dynamic Provisioner and FlexVolume
Examples for creation and removal of Ubiquity volumes (PV and PVC), as well as starting and stopping stateful containers\PODs are detailed in the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.

## Troubleshooting
### Log files
- The FlexVolume log name `ubiquity-k8s-flex.log`
- The Provisioner log name `ubiquity-k8s-provisioner.log`
- The Kubernetes kubelet logs can be viewed by running `journalctl -u kubelet"

## Support
For any questions, suggestions, or issues, use github.

## Licensing

Copyright 2016, 2017 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
