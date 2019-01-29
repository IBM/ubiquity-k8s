package watcher

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	v1client "k8s.io/client-go/kubernetes/typed/core/v1"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/IBM/ubiquity/utils/logs"
)

func GenerateSvcWatcher(name, namespace string, client v1client.ServicesGetter, logger logs.Logger) (watch.Interface, error) {
	logger.Info(fmt.Sprintf("Generating watcher for Service %s", name))
	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	}
	watcher, err := client.Services(namespace).Watch(listOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("Can't generate watcher for Service %s", name))
		return nil, err
	} else {
		logger.Info(fmt.Sprintf("Generated watcher for Service %s", name))
		return watcher, nil
	}
}
