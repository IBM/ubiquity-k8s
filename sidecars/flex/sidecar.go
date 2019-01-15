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

package flex

import (
	"context"
	"fmt"

	"github.com/IBM/ubiquity-k8s/utils"
	watcher "github.com/IBM/ubiquity-k8s/utils/watcher"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ServiceSyncer struct {
	name, namespace string
	kubeClient      kubernetes.Interface
	ctx             context.Context
	handler         cache.ResourceEventHandler
}

func NewServiceSyncer(kubeClient kubernetes.Interface, ctx context.Context) (*ServiceSyncer, error) {
	ns, err := utils.GetCurrentNamespace()
	if err != nil {
		return nil, err
	}

	ss := &ServiceSyncer{
		name:       utils.UbiquityServiceName,
		namespace:  ns,
		kubeClient: kubeClient,
		ctx:        ctx,
	}

	h := cache.ResourceEventHandlerFuncs{
		AddFunc:    ss.processService,
		UpdateFunc: ss.processServiceUpdate,
	}
	ss.handler = h
	return ss, nil
}

// Sync watches the ubiquity service and sync its CLusterIP changes to flex config file
func (ss *ServiceSyncer) Sync() error {
	ubiquitySvcWatcher, err := watcher.GenerateSvcWatcher(
		ss.name, ss.namespace,
		ss.kubeClient.CoreV1(),
		logger)
	if err != nil {
		return err
	}

	// process service for the first time if it is already existing.
	svc, err := ss.kubeClient.CoreV1().Services(ss.namespace).Get(ss.name, metav1.GetOptions{})
	if err == nil {
		ss.processService(svc)
	}
	err = watcher.Watch(ubiquitySvcWatcher, ss.handler, ss.ctx, logger)
	return err
}

// processService compare the ubiquity IP between service and config file and apply
// the new value to flex config file if they are different.
func (ss *ServiceSyncer) processService(obj interface{}) {
	currentFlexConfig, err := defaultFlexConfigSyncer.GetCurrentFlexConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("Can't read flex config file: %v", err))
		return
	}
	ubiquityIP := currentFlexConfig.UbiquityServer.Address
	svc := obj.(*v1.Service)
	if svc != nil && svc.Spec.ClusterIP != ubiquityIP {
		currentFlexConfig.UbiquityServer.Address = svc.Spec.ClusterIP
		err := defaultFlexConfigSyncer.UpdateFlexConfig(currentFlexConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("Can't write flex config file: %v", err))
			return
		}
	}
}

func (ss *ServiceSyncer) processServiceUpdate(old, cur interface{}) {
	if old == nil {
		ss.processService(cur)
		return
	}
	oldSvc := old.(*v1.Service)
	curSvc := cur.(*v1.Service)
	if oldSvc.Spec.ClusterIP != curSvc.Spec.ClusterIP {
		ss.processService(cur)
	}
}
