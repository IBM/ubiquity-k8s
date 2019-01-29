package hookexecutor

import (
	"fmt"
	"os"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/IBM/ubiquity-k8s/utils"
	uberrors "github.com/IBM/ubiquity-k8s/utils/errors"
)

type preDeleteExecutor struct {
	*baseExcutor
}

func newPreDeleteExecutor(
	kubeClient kubernetes.Interface,
) *preDeleteExecutor {
	return &preDeleteExecutor{
		baseExcutor: &baseExcutor{
			kubeClient: kubeClient,
		},
	}
}

func (e *preDeleteExecutor) Execute() error {
	logger.Info("Performing actions in pre-delete")
	var err error

	err = e.deleteUbiquityDBPods()
	if err != nil {
		return logger.ErrorRet(err, "Failed performing actions in pre-delete")
	}

	err = e.deleteUbiquityDBPvc()
	if err != nil {
		return logger.ErrorRet(err, "Failed performing actions in pre-delete")
	}

	logger.Info("Successfully performed actions in pre-delete")
	return nil
}

// deleteUbiquityDBPods sets replicas of ubiquity-db deployment to 0 and
// wait for all the relevant pods to be deleted.
func (e *preDeleteExecutor) deleteUbiquityDBPods() error {
	ns, err := utils.GetCurrentNamespace()
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Setting replicas to zero to stop the %s pod in namespace %s", ubiquityDBDeploymentName, ns))

	deploy, err := e.kubeClient.AppsV1().Deployments(ns).Get(ubiquityDBDeploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("The Ubiquity DB Deployment is already deleted")
			return nil
		}
		return logger.ErrorRet(err, fmt.Sprintf("Failed to get Deployment %s from API Server", ubiquityDBDeploymentName))
	}

	if watchers, err := generatePodsWatchersInDeployment(deploy, e.kubeClient.CoreV1()); err != nil {
		return logger.ErrorRet(err, "Failed generating Pod watcher")
	} else {
		var wg sync.WaitGroup
		var watcherErr error

		// Watch and wait for pods to be deleted.
		for _, w := range watchers {
			wg.Add(1)
			go func(watcher watch.Interface) {
				_, err := Watch(watcher, nil, ubiquityDBPodDeletionTimeoutSecond*time.Second)
				if err != nil {
					if watcherErr == nil {
						watcherErr = err
					}
				}
				wg.Done()
			}(w)
		}

		// start the watcher first and then update the replicas.
		var zero = int32(0)
		deploy.Spec.Replicas = &zero
		_, err = e.kubeClient.AppsV1().Deployments(ns).Update(deploy)
		if err != nil {
			return logger.ErrorRet(err, "Failed deleting Ubiquity DB Pods")
		}

		wg.Wait()

		if watcherErr != nil {
			return logger.ErrorRet(err, "Failed waiting Pod to be deleted")
		} else {
			logger.Info(fmt.Sprintf("Successfully Deleted Pods under Ubiquity DB Deployment %s", ubiquityDBDeploymentName))
			return nil
		}
	}
}

// deleteUbiquityDBPvc deletes the ubiquity-db pvc and wait for pvc/pv to be deleted.
func (e *preDeleteExecutor) deleteUbiquityDBPvc() error {
	ns, err := utils.GetCurrentNamespace()
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Deleting PVC %s in namespace %s", ubiquityDBPvcName, ns))

	pvc, err := e.kubeClient.CoreV1().PersistentVolumeClaims(ns).Get(ubiquityDBPvcName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("The Ubiquity DB PVC %s is already deleted", ubiquityDBPvcName))
			// if the pvc is already deleted, maybe it is a retry, we should not change
			// anything, just wait for the pv to be deleted.
			return e.waitUbiquityDbPvDeletion()
		}
		return logger.ErrorRet(err, fmt.Sprintf("Failed deleting Ubiquity DB PVC %s", ubiquityDBPvcName))
	}

	pvcWatcher, err := generatePvcWatcher(pvc.Name, ns, e.kubeClient.CoreV1())
	if err != nil {
		return logger.ErrorRet(err, "Failed generating PVC watcher")
	}
	pvWatcher, err := generatePvWatcher(pvc.Spec.VolumeName, e.kubeClient.CoreV1())
	if err != nil {
		return logger.ErrorRet(err, "Failed generating PV watcher")
	}

	var wg sync.WaitGroup
	var watcherErr error

	for _, w := range []watch.Interface{pvcWatcher, pvWatcher} {
		wg.Add(1)
		go func(watcher watch.Interface) {
			_, err := Watch(watcher, nil, ubiquityDBPvcAndPvDeletionTimeoutSecond*time.Second)
			if err != nil {
				if watcherErr == nil {
					watcherErr = err
				}
			}
			wg.Done()
		}(w)
	}

	// start the watcher first and then delete the resource
	if err := e.kubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(ubiquityDBPvcName, nil); err != nil {
		return logger.ErrorRet(err, fmt.Sprintf("Failed deleting Ubiquity DB PVC %s", ubiquityDBPvcName))
	}

	wg.Wait()

	if watcherErr != nil {
		return logger.ErrorRet(watcherErr, "Failed waiting PVC/PV to be deleted")
	} else {
		logger.Info(fmt.Sprintf("Successfully Deleted PVC %s", ubiquityDBPvcName))
		return nil
	}

}

func (e *preDeleteExecutor) waitUbiquityDbPvDeletion() error {
	logger.Info("Waiting for ubiquity DB PV to be deleted")
	name, err := getUbiquityDbPvName()
	if err != nil {
		return err
	}

	_, err = e.kubeClient.CoreV1().PersistentVolumes().Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Warning(fmt.Sprintf("The Ubiquity DB PV %s is already deleted", name))
			return nil
		}
		return logger.ErrorRet(err, "Failed waiting for ubiquity DB PV to be deleted")
	}

	pvWatcher, err := generatePvWatcher(name, e.kubeClient.CoreV1())
	if err != nil {
		return logger.ErrorRet(err, "Failed generating PV watcher")
	}
	_, err = Watch(pvWatcher, nil)
	return err
}

func getUbiquityDbPvName() (string, error) {
	name := os.Getenv("UBIQUITY_DB_PV_NAME")
	if name == "" {
		return "", uberrors.ENVUbiquityDbPvNameNotSet
	}
	return name, nil
}
