/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logging

import (
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/instrumentation"

	. "github.com/onsi/ginkgo"
)

const (
	// eventsIngestionTimeout is the amount of time to wait until some
	// events are ingested.
	eventsIngestionTimeout = 10 * time.Minute

	// eventPollingInterval is the delay between attempts to read events
	// from the logs provider.
	eventPollingInterval = 1 * time.Second

	// eventCreationInterval is the minimal delay between two events
	// created for testing purposes.
	eventCreationInterval = 10 * time.Second
)

var _ = instrumentation.SIGDescribe("Cluster level logging implemented by Stackdriver", func() {
	f := framework.NewDefaultFramework("sd-logging-events")

	BeforeEach(func() {
		framework.SkipUnlessProviderIs("gce", "gke")
	})

	It("should ingest events", func() {
		sdLogsProvider, err := newSdLogsProvider(f)
		framework.ExpectNoError(err, "Failed to create Stackdriver logs provider")

		err = sdLogsProvider.Init()
		defer sdLogsProvider.Cleanup()
		framework.ExpectNoError(err, "Failed to init Stackdriver logs provider")

		stopCh := make(chan struct{})
		successCh := make(chan struct{})
		go func() {
			wait.Poll(eventPollingInterval, eventsIngestionTimeout, func() (bool, error) {
				events := sdLogsProvider.ReadEvents()
				if len(events) > 0 {
					framework.Logf("Some events are ingested, sample event: %v", events[0])
					close(successCh)
					return true, nil
				}
				return false, nil
			})
			close(stopCh)
		}()

		By("Running pods to generate events while waiting for some of them to be ingested")
		wait.PollUntil(eventCreationInterval, func() (bool, error) {
			podName := "synthlogger"
			startNewLoggingPod(f, podName, "", 1, 1*time.Second)
			defer f.PodClient().Delete(podName, &meta_v1.DeleteOptions{})
			err = framework.WaitForPodSuccessInNamespace(f.ClientSet, podName, f.Namespace.Name)
			if err != nil {
				framework.Logf("Failed to wait pod %s to successfully complete due to %v", podName, err)
			}

			return false, nil
		}, stopCh)

		select {
		case <-successCh:
			break
		default:
			framework.Failf("No events are present in Stackdriver after %v", eventsIngestionTimeout)
		}
	})
})
