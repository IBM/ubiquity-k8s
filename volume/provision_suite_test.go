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

package volume_test

import (
	"fmt"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var testLogger *log.Logger
var logFile *os.File

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provisioner Suite")
}

var _ = BeforeEach(func() {
	var err error
	logFile, err = os.OpenFile("/tmp/test-ubiquity-provisioner.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return
	}
	testLogger = log.New(logFile, "provisioner: ", log.Lshortfile|log.LstdFlags)
})

var _ = AfterEach(func() {
	err := logFile.Sync()
	if err != nil {
		panic(err.Error())
	}
	err = logFile.Close()
	if err != nil {
		panic(err.Error())
	}
})
