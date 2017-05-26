package main

import (
	"flag"

	"time"

	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/IBM/ubiquity-k8s/volume"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/leaderelection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	provisioner          = flag.String("provisioner", "ubiquity/spectrum-scale", "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
	master               = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig           = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	configFile           = flag.String("config", "ubiquity-client.conf", "config file with ubiquity client configuration params")
	failedRetryThreshold = flag.Int("retries", 3, "number of retries on failure of provisioner")
)

const (
	leasePeriod   = leaderelection.DefaultLeaseDuration
	retryPeriod   = leaderelection.DefaultRetryPeriod
	renewDeadline = leaderelection.DefaultRenewDeadline
	termLimit     = leaderelection.DefaultTermLimit
)

func main() {

	flag.Parse()
	var ubiquityConfig resources.UbiquityPluginConfig
	fmt.Printf("Starting ubiquity plugin with %s config file\n", *configFile)
	if _, err := toml.DecodeFile(*configFile, &ubiquityConfig); err != nil {
		fmt.Println(err)
		return
	}
	logger, logFile := utils.SetupLogger(ubiquityConfig.LogPath, "ubiquity-provisioner")
	defer utils.CloseLogs(logFile)

	logger.Printf("Provisioner %s specified", *provisioner)

	// Create the client according to whether we are running in or out-of-cluster
	var config *rest.Config
	var err error
	if *master != "" || *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
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
		logger.Printf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	// nfsProvisioner := vol.NewNFProvisioner(exportDir, clientset, *useGanesha, ganeshaConfig)
	flexProvisioner, err := volume.NewFlexProvisioner(logger, remoteClient, ubiquityConfig)
	if err != nil {
		panic("Error starting ubiquity client")
	}
	// Start the provision controller which will dynamically provision NFS PVs

	pc := controller.NewProvisionController(clientset, 15*time.Second, *provisioner, flexProvisioner, serverVersion.GitVersion, true, *failedRetryThreshold, leasePeriod, renewDeadline, retryPeriod, termLimit)
	pc.Run(wait.NeverStop)
}
