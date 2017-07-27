# Ubiquity-k8s
This projects contains the needed components to manage persistent storage for kubernetes through [Ubiquity](https://github.com/IBM/ubiquity) service.
The repository contains two components that could be used separately or combined according to the requirements:

- [Ubiquity Dynamic Provisioner](ubiquity-dynamic-provisioner)
- [Ubiquity Flex Driver CLI](ubiquity-flexvolume-cli)

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Ubiquity Dynamic Provisioner 

Ubiquity Dynamic Provisioner facilitates creation and deletion of persistent volumes in Kubernetes (version 1.5.6) through use of the [Ubiquity](https://github.com/IBM/ubiquity) service.
  
### Installing the Ubiquity Dynamic Provisioner
Install and configure the Provisioner only on one node in the Kubernetes cluster(minion or master).

### 1. Prerequisites
  * The Provisioner is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

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

### 2. Downloading and installing the Provisioner

* Download and unpack the application package.
     ```bash
         mkdir -p /etc/ubiquity
         cd /etc/ubiquity
         curl -L https://github.com/IBM/ubiquity-k8s/releases/download/v0.4.0/ubiquity-k8s-provisioner-0.4.0.tar.gz | tar xf -
         cp provisioner /usr/bin 
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
Before running the Provisioner service, you must create and configure the `/etc/ubiquity/ubiquity-k8s-provisioner.conf` file, according to your storage system type.

Here is example of a configuration file that need to be set:
```toml
logPath = "/tmp/ubiquity"  # The Ubiquity provisioner will write logs to file "ubiquity-provisioner.log" in this path.
backend = "scbe" # Backend name such as scbe or spectrum-scale

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening
```

### 4. Running the Provisioner service
  * Run the service.
```bash
systemctl start ubiquity-k8s-provisioner    
```

### Provisioner usage examples
For examples on how to create and remove Ubiquity volumes(PV and PVC) refer to the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.


## Ubiquity FlexVolume CLI 

Ubiquity FlexVolume CLI supports attaching and detaching volumes on persistent storage using the [Ubiquity](https://github.com/IBM/ubiquity) service.

### Installing the Ubiquity FlexVolume
Install and configure the plugin on each node(minion) in the Kubernetes cluster that requires access to Ubiquity volumes.

### 1. Prerequisites
  * Ubiquity FlexVolume is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

  * Ubiquity FlexVolume requires Kubernetes version 1.5.6.

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
   

### 2. Downloading and installing the Ubiquity FlexVolume

* Download and unpack the application package.
     ```bash
         mkdir -p /usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity/ubiquity
         cd $_
         curl -L https://github.com/IBM/ubiquity-k8s/releases/download/v0.4.0/ubiquity-k8s-flex
         chmod u+x ubiquity-k8s-flex
         #chown USER:GROUP ubiquity-k8s-flex   ### Run this command only a non-root user.
     ```

### 3. Configuring the Ubiquity FlexVolume
Before running the FlexVolume CLI, you must create and configure the `/etc/ubiquity/ubiquity-client.conf` file, according to your storage system type.
Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section.

Here is example of a generic configuration file that need to be set:
```toml
logPath = "/tmp/ubiquity"  # The Ubiquity provisioner will write logs to file "ubiquity-provisioner.log" in this path.
backend = "scbe" # Backend name such as scbe or spectrum-scale

[UbiquityServer]
address = "127.0.0.1"  # IP/host of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening
```

### FlexVolume usage examples
For examples on how to start and stop stateful containers\PODs with Ubiquity volumes , refer to the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.


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
