/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	k8sutils "github.com/IBM/ubiquity-k8s/utils"
	"github.com/IBM/ubiquity-k8s/volume"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/utils"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"flag"
)

var (
	provisioner = k8sresources.ProvisionerName
	configFile  = os.Getenv("KUBECONFIG")
)

func main() {

	flag.CommandLine.Parse([]string{})

	ubiquityConfig, err := k8sutils.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to load config %#v", err))
	}
	fmt.Printf("Starting ubiquity plugin with %s config file\n", configFile)

	err = os.MkdirAll(ubiquityConfig.LogPath, 0640)
	if err != nil {
		panic(fmt.Errorf("Failed to setup log dir"))
	}

	defer k8sutils.InitProvisionerLogger(ubiquityConfig)()
	logger := utils.SetupOldLogger(k8sresources.UbiquityProvisionerName)

	logger.Printf("Provisioner %s specified", provisioner)

	var config *rest.Config

	if configFile != "" {
		logger.Printf("Uses k8s configuration file name %s", configFile)
		config, err = clientcmd.BuildConfigFromFlags("", configFile)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to create config: %v", err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		panic(fmt.Sprintf("Error getting server version: %v", err))
	}
	remoteClient, err := remote.NewRemoteClientSecure(logger, ubiquityConfig)
	if err != nil {
		logger.Printf("Error getting remote Client: %v", err)
		panic("Error getting remote client")
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	ubiquityConfigCopyWithPasswordStarred := ubiquityConfig
	ubiquityConfigCopyWithPasswordStarred.CredentialInfo.Password = "****"
	logger.Printf("starting the provisioner, remote client %#v, config %#v", remoteClient, ubiquityConfigCopyWithPasswordStarred)
	flexProvisioner, err := volume.NewFlexProvisioner(logger, remoteClient, ubiquityConfig)
	if err != nil {
		logger.Printf("Error starting provisioner: %v", err)
		panic("Error starting ubiquity client")
	}

	// Start the provision controller which will dynamically provision Ubiquity PVs

	pc := controller.NewProvisionController(clientset, provisioner, flexProvisioner, serverVersion.GitVersion)
	pc.Run(wait.NeverStop)
}
