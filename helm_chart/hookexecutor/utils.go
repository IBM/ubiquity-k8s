package hookexecutor

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	v1client "k8s.io/client-go/kubernetes/typed/core/v1"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/IBM/ubiquity/utils/logs"
)

const (
	defaultNamespace             = "ubiquity"
	ubiquityServiceName          = "ubiquity"
	ubiquityK8sFlexDaemonSetName = "ubiquity-k8s-flex"
	ubiquityDBDeploymentName     = "ubiquity-db"
	ubiquityDBPvcName            = "ibm-ubiquity-db"
	ubiquityK8sFlexContainerName = ubiquityK8sFlexDaemonSetName
	ubiquityIPAddressKey         = "UBIQUITY_IP_ADDRESS"
)

var logger = logs.GetLogger()

func match(ls *metav1.LabelSelector, pod *v1.Pod) bool {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return false
	}
	return selector.Matches(labels.Set(pod.Labels))
}

func generatePodsWatchersInDeployment(deploy *appsv1.Deployment, client v1client.PodsGetter) ([]watch.Interface, error) {
	pods, err := client.Pods(deploy.Namespace).List(metav1.ListOptions{})
	if err != nil {
		logger.Error("Can't get Pods from API Server")
		return nil, err
	}
	ls := deploy.Spec.Selector
	watchers := []watch.Interface{}

	for _, pod := range pods.Items {
		if !match(ls, &pod) {
			continue
		}

		logger.Info(fmt.Sprintf("Generating watcher for Pod %s", pod.Name))
		listOptions := metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, pod.Name).String(),
		}
		watcher, err := client.Pods(deploy.Namespace).Watch(listOptions)
		if err != nil {
			logger.Error(fmt.Sprintf("Can't generate watcher for Pod %s", pod.Name))
			return nil, err
		} else {
			logger.Info(fmt.Sprintf("Generated watcher for Pod %s", pod.Name))
			watchers = append(watchers, watcher)
		}
	}
	return watchers, nil
}

func generatePodWatcher(name, namespace string, client v1client.PodsGetter) (watch.Interface, error) {
	logger.Info(fmt.Sprintf("Generating watcher for Pod %s", name))
	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	}
	watcher, err := client.Pods(namespace).Watch(listOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("Can't generate watcher for Pod %s", name))
		return nil, err
	} else {
		logger.Info(fmt.Sprintf("Generated watcher for Pod %s", name))
		return watcher, nil
	}
}

func generatePvcWatcher(name, namespace string, client v1client.PersistentVolumeClaimsGetter) (watch.Interface, error) {
	logger.Info(fmt.Sprintf("Generating watcher for PVC %s", name))
	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	}
	watcher, err := client.PersistentVolumeClaims(namespace).Watch(listOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("Can't generate watcher for PVC %s", name))
		return nil, err
	} else {
		logger.Info(fmt.Sprintf("Generated watcher for PVC %s", name))
		return watcher, nil
	}
}

func generatePvWatcher(name string, client v1client.PersistentVolumesGetter) (watch.Interface, error) {
	logger.Info(fmt.Sprintf("Generating watcher for PV %s", name))
	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	}
	watcher, err := client.PersistentVolumes().Watch(listOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("Can't generate watcher for PV %s", name))
		return nil, err
	} else {
		logger.Info(fmt.Sprintf("Generated watcher for PV %s", name))
		return watcher, nil
	}
}

// desiredStateChecker: check if the resource meet the desired state, if it is nil,
// means that the caller expects the resource to be deleted.
func Watch(watcher watch.Interface, desiredStateChecker func(runtime.Object) bool, timeout ...time.Duration) (bool, error) {
	var err error

	t := 30 * time.Second
	if len(timeout) > 0 {
		t = timeout[0]
	}
	timer := time.NewTimer(t)

	logger.Info("Start watching resource")

outerLoop:
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Modified {
				metadata, err := meta.Accessor(event.Object)
				if err != nil {
					logger.Error("Can not get resource metadata")
				} else {
					if metadata.GetDeletionTimestamp() != nil {
						logger.Info("The resource is terminating")
						if desiredStateChecker != nil {
							// the resource is terminating, it can not reach the desired state any more.
							err = fmt.Errorf("The resource is terminating")
							break outerLoop
						}
					}
				}

				if desiredStateChecker != nil {
					if desiredStateChecker(event.Object) {
						logger.Info("The resource reaches the desired state.")
						break outerLoop
					}

				}
			}
			if event.Type == watch.Deleted {
				logger.Info("The resource is deleted.")
				break outerLoop
			}
		case <-timer.C:
			err = fmt.Errorf("Timed out waiting for new resource changes")
			logger.Error(err.Error())
			break outerLoop
		}
	}
	// stop watching this resource
	watcher.Stop()
	logger.Info("Stop watching resource")
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
