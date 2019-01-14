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
	"os"

	"github.com/IBM/ubiquity-k8s/utils"
	watcher "github.com/IBM/ubiquity-k8s/utils/watcher"
	"github.com/IBM/ubiquity/utils/logs"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var logger logs.Logger

func init() {
	initLogger()
}

func initLogger() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = utils.DefaultlogLevel
	}
	utils.InitGenericLogger(logLevel)
	logger = logs.GetLogger()
}

func SyncService(kubeClient kubernetes.Interface, ctx context.Context) error {
	ns, err := utils.GetCurrentNamespace()
	if err != nil {
		return err
	}

	ubiquitySvcWatcher, err := watcher.GenerateSvcWatcher(
		utils.UbiquityServiceName, ns,
		kubeClient.CoreV1(),
		logger)
	if err != nil {
		return err
	}

	h := cache.ResourceEventHandlerFuncs{
		AddFunc:    processService,
		UpdateFunc: processServiceUpdate,
	}

	// process service for the first time if it is already existing.
	svc, err := kubeClient.CoreV1().Services(ns).Get(utils.UbiquityServiceName, metav1.GetOptions{})
	if err == nil {
		processService(svc)
	}
	err = watcher.Watch(ubiquitySvcWatcher, h, ctx, logger)
	return err
}

func processService(obj interface{}) {
	currentFlexConfig, err := getCurrentFlexConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("Can't read flex config file: %v", err))
		return
	}
	ubiquityIP := currentFlexConfig.UbiquityServer.Address
	svc := obj.(*v1.Service)
	if svc != nil && svc.Spec.ClusterIP != ubiquityIP {
		currentFlexConfig.UbiquityServer.Address = svc.Spec.ClusterIP
		err := updateFlexConfig(currentFlexConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("Can't write flex config file: %v", err))
			return
		}
	}
}

func processServiceUpdate(old, cur interface{}) {
	if old == nil {
		processService(cur)
	}
	oldSvc := old.(*v1.Service)
	curSvc := cur.(*v1.Service)
	if oldSvc.Spec.ClusterIP != curSvc.Spec.ClusterIP {
		processService(cur)
	}
}
