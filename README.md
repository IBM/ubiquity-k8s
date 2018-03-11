# Ubiquity Kubernetes Persistent Storage
This project includes components for managing [Kubernetes persistent storage](https://kubernetes.io/docs/concepts/storage/persistent-volumes), using [Ubiquity](https://github.com/IBM/ubiquity) service.
- Ubiquity Dynamic Provisioner for creating and deleting persistent volumes
- Ubiquity FlexVolume Driver CLI for attaching and detaching persistent volumes

Currently, the following storage systems use Ubiquity:
* IBM block storage.

   The IBM block storage is supported for Kubernetes via IBM Spectrum Connect (3.4.0), previously known as IBM Spectrum Control Base Edition (3.3.0). Ubiquity communicates with the IBM storage systems through Spectrum Connect. Spectrum Connect creates a storage profile (for example, gold, silver or bronze) and makes it available for Kubernetes. For details about supported storage systems, refer to the latest Spectrum Connect release notes.
   
   The IBM official solution for Kubernetes, based on the Ubiquity project, is referred to as IBM Storage Enabler for Containers. You can download the installation package and its documentation from [IBM Fix Central](https://www-945.ibm.com/support/fixcentral/swg/selectFixes?parent=Software%2Bdefined%2Bstorage&product=ibm/StorageSoftware/IBM+Spectrum+Connect&release=All&platform=Linux&function=all). For details on the IBM Storage Enabler for Containers, see the relevant sections in the Spectrum Connect user guide.

* IBM Spectrum Scale, for testing only.

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.

## Solution overview

![Ubiquity Overview](images/ubiquity_architecture_draft_for_github.jpg)

Deployment description:
   *   Ubiquity Kubernetes Dynamic Provisioner (ubiquity-k8s-provisioner) runs as a Kubernetes deployment with replica=1.
   *   Ubiquity Kubernetes FlexVolume (ubiquity-k8s-flex) runs as a Kubernetes daemonset on all the worker and master nodes.
   *   Ubiquity (ubiquity) runs as a Kubernetes deployment with replica=1.
   *   Ubiquity database (ubiquity-db) runs as a Kubernetes deployment with replica=1.

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
