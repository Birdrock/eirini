package kubelet_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubelet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubelet Suite")
}
