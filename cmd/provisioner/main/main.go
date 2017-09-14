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
	"time"

	"fmt"
	"path"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	k8sutils "github.com/IBM/ubiquity-k8s/utils"
	"github.com/IBM/ubiquity-k8s/volume"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/leaderelection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strconv"
	"strings"
)

var (
	provisioner          = k8sresources.ProvisionerName
	configFile           = os.Getenv("KUBECONFIG")
	failedRetryThreshold = os.Getenv("RETRIES")
)

const (
	leasePeriod   = leaderelection.DefaultLeaseDuration
	retryPeriod   = leaderelection.DefaultRetryPeriod
	renewDeadline = leaderelection.DefaultRenewDeadline
	termLimit     = leaderelection.DefaultTermLimit
)

func main() {

	ubiquityConfig, err := k8sutils.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to load config %#v", err))
	}
	fmt.Printf("Starting ubiquity plugin with %s config file\n", configFile)

	defer logs.InitFileLogger(logs.GetLogLevelFromString(ubiquityConfig.LogLevel), path.Join(ubiquityConfig.LogPath, k8sresources.UbiquityProvisionerLogFileName))()
	logger, logFile := utils.SetupLogger(ubiquityConfig.LogPath, k8sresources.UbiquityProvisionerName)
	defer utils.CloseLogs(logFile)

	logger.Printf("Provisioner %s specified", provisioner)

	// Create the client according to whether we are running in or out-of-cluster
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
	ubiquityEndpoint := fmt.Sprintf("http://%s:%d/ubiquity_storage", ubiquityConfig.UbiquityServer.Address, ubiquityConfig.UbiquityServer.Port)
	logger.Printf("ubiquity endpoint")
	remoteClient, err := remote.NewRemoteClient(logger, ubiquityEndpoint, ubiquityConfig)
	if err != nil {
		logger.Printf("Error getting remote Client: %v", err)
		panic("Error getting remote client")
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	// nfsProvisioner := vol.NewNFProvisioner(exportDir, clientset, *useGanesha, ganeshaConfig)
	fmt.Printf("starting the provisioner with logger %#v , remote client %#v and config %#v", logger, remoteClient, ubiquityConfig)
	flexProvisioner, err := volume.NewFlexProvisioner(logger, remoteClient, ubiquityConfig)
	if err != nil {
		logger.Printf("Error starting provisioner: %v", err)
		panic("Error starting ubiquity client")
	}
	intVal, err := strconv.ParseInt(failedRetryThreshold, 0, 32)
	if err != nil {
		logger.Printf("Error parsing retries: %v", err)
		panic("Error getting retries value")
	}
	failedRetryThresholdInt := int(intVal)
	// Start the provision controller which will dynamically provision NFS PVs

	pc := controller.NewProvisionController(clientset, 15*time.Second, provisioner, flexProvisioner, serverVersion.GitVersion, true, failedRetryThresholdInt, leasePeriod, renewDeadline, retryPeriod, termLimit)
	pc.Run(wait.NeverStop)
}

func LoadConfig() (resources.UbiquityPluginConfig, error) {

	config := resources.UbiquityPluginConfig{}
	config.LogLevel = os.Getenv("LOG_LEVEL")
	config.LogPath = os.Getenv("LOG_PATH")
	config.Backends = strings.Split(os.Getenv("BACKENDS"), ",")
	ubiquity := resources.UbiquityServerConnectionInfo{}
	port, err := strconv.ParseInt(os.Getenv("UBIQUITY_PORT"), 0, 32)
	if err != nil {
		return config, err
	}
	ubiquity.Port = int(port)
	ubiquity.Address = os.Getenv("UBIQUITY_ADDRESS")
	config.UbiquityServer = ubiquity
	return config, nil
}
