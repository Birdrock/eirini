package eats_test

import (
	"os"

	"code.cloudfoundry.org/eirini/integration/util"
	"code.cloudfoundry.org/eirini/models/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Staging", func() {

	const stagingGUID = "staging-guid-1"
	var (
		certPath              string
		keyPath               string
		cloudControllerServer *ghttp.Server
	)

	BeforeEach(func() {
		var err error
		certPath, keyPath = generateKeyPair("capi")
		cloudControllerServer, err = util.CreateTestServer(certPath, keyPath, certPath)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		cloudControllerServer.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
	})

	It("stages the application", func() {
		stagingRequest := cf.StagingRequest{}
		stageLRP(httpClient, opiURL, stagingGUID, stagingRequest)
	})

})
