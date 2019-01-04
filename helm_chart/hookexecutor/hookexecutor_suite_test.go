package hookexecutor

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHookexecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hookexecutor Suite")
}
