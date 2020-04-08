package eats_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
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
		opiConfig        string
		opiSession       *gexec.Session
		collectorSession *gexec.Session
		collectorConfig  string
		httpClient       *http.Client
		opiURL           string

		localhostCertPath, localhostKeyPath string

		natsConfig *server.Options
		natsServer *server.Server
		natsClient *nats.Conn

		registerChan   chan *nats.Msg
		unregisterChan chan *nats.Msg
	)

	BeforeEach(func() {
		localhostCertPath, localhostKeyPath = generateKeyPair("localhost")

		registerChan = make(chan *nats.Msg)
		unregisterChan = make(chan *nats.Msg)

		var err error
		httpClient, err = makeTestHTTPClient(localhostCertPath, localhostKeyPath)
		Expect(err).ToNot(HaveOccurred())

		opiSession, opiConfig, opiURL = runOpi(localhostCertPath, localhostKeyPath)
		waitOpiReady(httpClient, opiURL)

		natsConfig = getNatsServerConfig()
		natsServer = natstest.RunServer(natsConfig)
		natsClient = subscribeToNats(natsConfig, registerChan, unregisterChan)

		eiriniRouteConfig := eirini.RouteEmitterConfig{
			NatsPassword: natsConfig.Password,
			NatsIP:       natsConfig.Host,
			NatsPort:     natsConfig.Port,
			KubeConfig: eirini.KubeConfig{
				ConfigPath: fixture.KubeConfigPath,
				Namespace:  fixture.Namespace,
			},
		}
		collectorSession, collectorConfig = runBinary("code.cloudfoundry.org/eirini/cmd/route-collector", eiriniRouteConfig)
	})

	AfterEach(func() {
		if opiSession != nil {
			opiSession.Kill()
		}
		if collectorSession != nil {
			collectorSession.Kill()
		}
		if natsServer != nil {
			natsServer.Shutdown()
		}
		if natsClient != nil {
			natsClient.Close()
		}
		Expect(os.Remove(opiConfig)).To(Succeed())
		Expect(os.Remove(collectorConfig)).To(Succeed())
	})

	Context("When an app is running", func() {
		var lrp cf.DesireLRPRequest

		BeforeEach(func() {
			lrp = cf.DesireLRPRequest{
				GUID:         "the-app-guid",
				Version:      "0.0.0",
				NumInstances: 1,
				Routes: map[string]*json.RawMessage{
					"cf-router": marshalRoutes([]routeInfo{
						{Hostname: "app-hostname-1", Port: 8080},
						{Hostname: "app-hostname-2", Port: 9090},
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
			time.Sleep(20 * time.Second)
		})

		It("registers its routes", func() {
			var msg nats.Msg
			Eventually(registerChan, "1m").Should(Receive(&msg))
			fmt.Printf("Received msg: %#v\n", msg)
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

func subscribeToNats(natsConfig *server.Options, registerChan, unregisterChan chan<- *nats.Msg) *nats.Conn {
	natsClientConfig := nats.GetDefaultOptions()
	natsClientConfig.Timeout = (5 * time.Second)
	natsClientConfig.ReconnectWait = (20 * time.Millisecond)
	natsClientConfig.MaxReconnect = 1000
	natsClientConfig.NoRandomize = true
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
