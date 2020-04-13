package eats_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/route"
	"github.com/nats-io/gnatsd/server"
	natstest "github.com/nats-io/nats-server/test"
	"github.com/nats-io/nats.go"
	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
)

type routeInfo struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

var _ = FDescribe("Routes", func() {

	var (
		collectorSession   *gexec.Session
		collectorConfig    string
		uriInformerSession *gexec.Session
		uriInformerConfig  string

		natsConfig *server.Options
		natsServer *server.Server
		natsClient *nats.Conn

		registerChan   chan *nats.Msg
		unregisterChan chan *nats.Msg

		lrp cf.DesireLRPRequest
	)

	BeforeEach(func() {
		registerChan = make(chan *nats.Msg)
		unregisterChan = make(chan *nats.Msg)

		natsConfig = getNatsServerConfig()
		natsServer = natstest.RunServer(natsConfig)
		natsClient = subscribeToNats(natsConfig, registerChan, unregisterChan)

		eiriniRouteConfig := eirini.RouteEmitterConfig{
			NatsPassword:        natsConfig.Password,
			NatsIP:              natsConfig.Host,
			NatsPort:            natsConfig.Port,
			EmitPeriodInSeconds: 1,
			KubeConfig: eirini.KubeConfig{
				ConfigPath: fixture.KubeConfigPath,
				Namespace:  fixture.Namespace,
			},
		}
		collectorSession, collectorConfig = runBinary("code.cloudfoundry.org/eirini/cmd/route-collector", eiriniRouteConfig)
		uriInformerSession, uriInformerConfig = runBinary("code.cloudfoundry.org/eirini/cmd/route-statefulset-informer", eiriniRouteConfig)

		lrp = cf.DesireLRPRequest{
			GUID:         "the-app-guid",
			Version:      "the-version",
			NumInstances: 1,
			Routes: map[string]*json.RawMessage{
				"cf-router": marshalRoutes([]routeInfo{
					{Hostname: "app-hostname-1", Port: 8080},
				}),
			},
			Ports: []int32{8080},
			Lifecycle: cf.Lifecycle{
				DockerLifecycle: &cf.DockerLifecycle{
					Image: "eirini/dorini",
				},
			},
		}
	})

	JustBeforeEach(func() {
		resp, err := desireLRP(httpClient, opiURL, lrp)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
	})

	AfterEach(func() {
		if collectorSession != nil {
			collectorSession.Kill()
		}
		if uriInformerSession != nil {
			uriInformerSession.Kill()
		}
		if natsServer != nil {
			natsServer.Shutdown()
		}
		if natsClient != nil {
			natsClient.Close()
		}
		Expect(os.Remove(collectorConfig)).To(Succeed())
		Expect(os.Remove(uriInformerConfig)).To(Succeed())
	})

	It("continuously registers its routes", func() {
		var msg *nats.Msg

		for i := 0; i < 5; i++ {
			Eventually(registerChan, "15s").Should(Receive(&msg))
			var actualMessage route.RegistryMessage
			Expect(json.Unmarshal(msg.Data, &actualMessage)).To(Succeed())
			Expect(net.ParseIP(actualMessage.Host).IsUnspecified()).To(BeFalse())
			Expect(actualMessage.Port).To(BeNumerically("==", 8080))
			Expect(actualMessage.URIs).To(ConsistOf("app-hostname-1"))
			Expect(actualMessage.App).To(Equal("the-app-guid"))
			Expect(actualMessage.PrivateInstanceID).To(ContainSubstring("the-app-guid"))
		}
	})

	When("the app fails to start", func() {
		BeforeEach(func() {
			lrp.Lifecycle.DockerLifecycle.Image = "eirini/does-not-exist"
		})

		It("does not register routes", func() {
			Consistently(registerChan, "5s").ShouldNot(Receive())
		})
	})

	When("a new route is added to the app", func() {

		JustBeforeEach(func() {
			Expect(updateLRP(httpClient, opiURL, cf.UpdateDesiredLRPRequest{
				UpdateDesiredLRPRequest: models.UpdateDesiredLRPRequest{
					ProcessGuid: "the-app-guid-the-version",
					Update: &models.DesiredLRPUpdate{
						Routes: &models.Routes{
							"cf-router": marshalRoutes([]routeInfo{
								{Hostname: "app-hostname-1", Port: 8080},
								{Hostname: "app-hostname-2", Port: 9090},
							}),
						},
					},
				},
			})).To(Succeed())

		})

		FIt("registers the new route", func() {
			var msg *nats.Msg
			Eventually(registerChan, "15s").Should(Receive(&msg))
			var actualMessage route.RegistryMessage
			Expect(json.Unmarshal(msg.Data, &actualMessage)).To(Succeed())
			Expect(actualMessage.URIs).To(ConsistOf("app-hostname-1", "app-hostname-2"))
		})
	})
})

func getNatsServerConfig() *server.Options {
	return &server.Options{
		Host:           "127.0.0.1",
		Port:           51000 + rand.Intn(1000) + ginkgoconfig.GinkgoConfig.ParallelNode,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 2048,
		Username:       "nats",
		Password:       "s3cr3t",
	}
}

func marshalRoutes(routes []routeInfo) *json.RawMessage {
	bytes, err := json.Marshal(routes)
	Expect(err).NotTo(HaveOccurred())

	rawMessage := &json.RawMessage{}
	Expect(rawMessage.UnmarshalJSON(bytes)).To(Succeed())
	return rawMessage
}

func runBinary(binPath string, config interface{}) (*gexec.Session, string) {
	binaryPath, err := gexec.Build(binPath)
	Expect(err).NotTo(HaveOccurred())

	configBytes, err := yaml.Marshal(config)
	Expect(err).NotTo(HaveOccurred())

	configFile := writeTempFile(configBytes, filepath.Base(binaryPath)+"-config.yaml")
	command := exec.Command(binaryPath, "-c", configFile) // #nosec G204
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session, configFile
}

func subscribeToNats(natsConfig *server.Options, registerChan, unregisterChan chan *nats.Msg) *nats.Conn {
	natsClientConfig := nats.GetDefaultOptions()
	natsClientConfig.Servers = []string{fmt.Sprintf("%s:%d", natsConfig.Host, natsConfig.Port)}
	natsClientConfig.User = natsConfig.Username
	natsClientConfig.Password = natsConfig.Password
	natsClient, err := natsClientConfig.Connect()
	Expect(err).ToNot(HaveOccurred())

	natsClient.Subscribe("router.register", func(msg *nats.Msg) {
		registerChan <- msg
	})
	natsClient.Subscribe("router.unregister", func(msg *nats.Msg) {
		unregisterChan <- msg
	})

	return natsClient
}
