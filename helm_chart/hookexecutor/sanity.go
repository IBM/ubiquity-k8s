package hookexecutor

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type sanityExecutor struct {
	*baseExcutor
}

func newSanityExecutor(
	kubeClient kubernetes.Interface,
) *sanityExecutor {
	return &sanityExecutor{
		baseExcutor: &baseExcutor{
			kubeClient: kubeClient,
		},
	}
}

func (e *sanityExecutor) Execute() error {
	logger.Info("Performing actions in sanity test")

	var err error
	msg := "Sanity test failed, please investigate the sanity resources with name sanity-* and delete them after that manually"
	err = e.createSanityResources()
	if err != nil {
		return logger.ErrorRet(err, msg)
	}

	err = e.deleteSanityResources()
	if err != nil {
		return logger.ErrorRet(err, msg)
	}

	logger.Info("Successfully performed actions in sanity test")
	return nil
}

// createSanityResources creates a pvc and a pod and wait utill they reach the desired state.
func (e *sanityExecutor) createSanityResources() error {
	logger.Info("Creating sanity resources")

	var err error
	pvc, pod := getSanityPvcAndPod()

	err = updateStorageClassInPvc(pvc)
	if err != nil {
		return logger.ErrorRet(err, "Failed to create sanity resources")
	}

	err = updateNamespace([]runtime.Object{pvc, pod})
	if err != nil {
		return logger.ErrorRet(err, "Failed to move the sanity resources to the specified namespace")
	}

	logger.Info("Creating sanity PVCs")
	_, err = e.kubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(pvc)
	if err != nil {
		return logger.ErrorRet(err, "Failed to create sanity pvcs")
	}

	pvcWatcher, err := generatePvcWatcher(pvc.Name, pvc.Namespace, e.kubeClient.CoreV1())
	if err != nil {
		return logger.ErrorRet(err, "Failed generating PVC watcher")
	}
	_, err = Watch(pvcWatcher, func(obj runtime.Object) bool {
		pvc := obj.(*corev1.PersistentVolumeClaim)
		return pvc.Status.Phase == corev1.ClaimBound
	})
	if err != nil {
		return logger.ErrorRet(err, "Failed waiting for PVC to be bound")
	}

	logger.Info("Creating sanity Pods")
	_, err = e.kubeClient.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		return logger.ErrorRet(err, "Failed to create sanity pods")
	}

	podWatcher, err := generatePodWatcher(pod.Name, pod.Namespace, e.kubeClient.CoreV1())
	if err != nil {
		return logger.ErrorRet(err, "Failed generating Pod watcher")
	}
	_, err = Watch(podWatcher, func(obj runtime.Object) bool {
		pod := obj.(*corev1.Pod)
		return pod.Status.Phase == corev1.PodRunning
	}, sanityPodRunningTimeoutSecond*time.Second)
	if err != nil {
		return logger.ErrorRet(err, "Failed waiting for Pod to be running")
	}

	logger.Info("Successfully created sanity resources")
	return nil
}

// deleteSanityResources deletes the pod and pvc and wait utill they are deleted.
func (e *sanityExecutor) deleteSanityResources() error {
	logger.Info("Deleting sanity resources")

	var err error
	pvc, pod := getSanityPvcAndPod()
	updateNamespace([]runtime.Object{pvc, pod})

	succeeded := make(chan bool)

	// watch sanity Pods and Pvcs to be deleted
	go func() {
		defer close(succeeded)

		podWatcher, _ := generatePodWatcher(pod.Name, pod.Namespace, e.kubeClient.CoreV1())
		_, err := Watch(podWatcher, nil, sanityPodDeletionTimeoutSecond*time.Second)
		if err != nil {
			logger.Error("Failed waiting for Pod to be deleted")
			succeeded <- false
		}
		succeeded <- true

		pvcWatcher, _ := generatePvcWatcher(pvc.Name, pvc.Namespace, e.kubeClient.CoreV1())
		_, err = Watch(pvcWatcher, nil, sanityPvcDeletionTimeoutSecond*time.Second)
		if err != nil {
			logger.Error("Failed waiting for PVC to be deleted")
			succeeded <- false
		}
		succeeded <- true
	}()

	logger.Info("Deleting sanity Pods")
	err = e.kubeClient.CoreV1().Pods(pod.Namespace).Delete(pod.Name, nil)
	if err != nil {
		return logger.ErrorRet(err, "Failed to delete sanity pods")
	}

	done := <-succeeded
	if !done {
		return fmt.Errorf("Failed waiting for Pod to be deleted")
	}

	logger.Info("Deleting sanity PVCs")
	err = e.kubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(pvc.Name, nil)
	if err != nil {
		return logger.ErrorRet(err, "Failed to delete sanity pvcs")
	}

	done = <-succeeded
	if !done {
		return fmt.Errorf("Failed waiting for PVC to be deleted")
	}

	logger.Info("Successfully deleted sanity resources")
	return nil
}

func getSanityPvcAndPod() (*corev1.PersistentVolumeClaim, *corev1.Pod) {
	podObj, _ := FromYaml([]byte(sanityPod))
	pod := podObj.(*corev1.Pod)
	pvcObj, _ := FromYaml([]byte(sanityPvc))
	pvc := pvcObj.(*corev1.PersistentVolumeClaim)
	return pvc, pod
}

func updateStorageClassInPvc(pvc *corev1.PersistentVolumeClaim) error {
	sc := os.Getenv("STORAGE_CLASS")
	if sc == "" {
		return fmt.Errorf(ENVStorageClassNotSet)
	}

	annos := pvc.GetAnnotations()
	if annos == nil {
		annos = make(map[string]string)
	}
	annos["volume.beta.kubernetes.io/storage-class"] = sc
	pvc.SetAnnotations(annos)
	return nil
}

func updateNamespace(objs []runtime.Object) error {
	ns, err := getCurrentNamespace()
	if err != nil {
		return err
	}
	for _, obj := range objs {
		metadata, err := meta.Accessor(obj)
		if err != nil {
			return logger.ErrorRet(err, "Failed to get metadata from resource")
		}
		logger.Info(fmt.Sprintf("Moving resource %s to namespace %s", metadata.GetName(), ns))
		metadata.SetNamespace(ns)
	}
	return nil
}
