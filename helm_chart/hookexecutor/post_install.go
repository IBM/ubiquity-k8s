package hookexecutor

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type postInstallExecutor struct {
	*baseExcutor
}

func newPostInstallExecutor(
	kubeClient kubernetes.Interface,
) *postInstallExecutor {
	return &postInstallExecutor{
		baseExcutor: &baseExcutor{
			kubeClient: kubeClient,
		},
	}
}

func (e *postInstallExecutor) Execute() error {
	logger.Info("Performing actions in post-install")
	err := e.updateFlexDaemonSet()
	if err != nil {
		return logger.ErrorRet(err, "Failed performing actions in post-install")
	} else {
		logger.Info("Successfully performed actions in post-install")
		return nil
	}
}

// updateFlexDaemonSet get the clusterIP of ubiquity Service and apply it to the flex DaemonSet
func (e *postInstallExecutor) updateFlexDaemonSet() error {
	logger.Info(fmt.Sprintf("Updating %s DaemonSet", ubiquityK8sFlexDaemonSetName))
	ns, err := getCurrentNamespace()
	if err != nil {
		return err
	}
	clusterIP, err := e.getUbiquityServiceIP()
	if err != nil {
		return logger.ErrorRet(err, fmt.Sprintf("Failed updating DaemonSet %s", ubiquityK8sFlexDaemonSetName))
	}

	logger.Info(
		fmt.Sprintf("Updating ubiquity serviceIP to %s in DaemonSet %s in namespace %s",
			clusterIP,
			ubiquityK8sFlexDaemonSetName,
			ns))

	flex, err := e.kubeClient.AppsV1().DaemonSets(ns).Get(ubiquityK8sFlexDaemonSetName, metav1.GetOptions{})
	if err != nil {
		return logger.ErrorRet(err, fmt.Sprintf("Failed getting DaemonSet %s", ubiquityK8sFlexDaemonSetName))
	}
	containers := flex.Spec.Template.Spec.Containers
	found := false
	// find the flex container and update the ubiquityIPAddress
	for i, container := range containers {
		if container.Name == ubiquityK8sFlexContainerName {
			envs := containers[i].Env
			for j, env := range envs {
				if env.Name == ubiquityIPAddressKey {
					envs[j].Value = clusterIP
					found = true
				}
			}
			if !found {
				logger.Warning(fmt.Sprintf("ENV %s is not found, creating it", ubiquityIPAddressKey))
				ipEnv := corev1.EnvVar{
					Name:  ubiquityIPAddressKey,
					Value: clusterIP,
				}
				newEnvs := append(envs, ipEnv)
				containers[i].Env = newEnvs
			}
		}
	}

	_, err = e.kubeClient.AppsV1().DaemonSets(ns).Update(flex)
	if err != nil {
		return logger.ErrorRet(err, fmt.Sprintf("Failed updating DaemonSet %s", ubiquityK8sFlexDaemonSetName))
	}
	logger.Info(fmt.Sprintf("Successfully updated DaemonSet %s", ubiquityK8sFlexDaemonSetName))
	return nil
}

func (e *postInstallExecutor) getUbiquityServiceIP() (string, error) {
	ns, err := getCurrentNamespace()
	if err != nil {
		return "", err
	}

	logger.Info(
		fmt.Sprintf("Getting ubiquity serviceIP from Service %s in namespace %s",
			ubiquityServiceName,
			ns))
	service, err := e.kubeClient.CoreV1().Services(ns).Get(ubiquityServiceName, metav1.GetOptions{})
	if err != nil {
		return "", logger.ErrorRet(err, "Failed getting ubiquity serviceIP")
	}
	ip := service.Spec.ClusterIP
	if ip == "" {
		err := fmt.Errorf(UbiquityServiceIPEmptyErrorStr)
		return "", logger.ErrorRet(err, err.Error())
	}
	return ip, nil
}

func getCurrentNamespace() (string, error) {
	ns := os.Getenv("NAMESPACE")
	if ns == "" {
		return "", fmt.Errorf(ENVNamespaceNotSet)
	}
	return ns, nil
}
