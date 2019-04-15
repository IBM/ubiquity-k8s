package flex

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFlex(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flex sidecar Suite")
}
