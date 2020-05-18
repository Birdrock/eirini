package opi_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/eirini"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Desire App", func() {
	var body string

	BeforeEach(func() {
		body = `{
			"guid": "the-app-guid",
			"version": "0.0.0",
			"ports": [8080],
			"instances": 2,
			"lifecycle": {
				"docker_lifecycle": {
					"image": "busybox",
					"command": ["sh", "-c", "env && sleep 1000"]
				}
			}
		}`
	})

	JustBeforeEach(func() {
		desireAppReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/apps/the-app-guid", url), bytes.NewReader([]byte(body)))
		Expect(err).NotTo(HaveOccurred())
		resp, err := httpClient.Do(desireAppReq)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
	})

	It("should create a stateful set for the app", func() {
		statefulsets, err := fixture.Clientset.AppsV1().StatefulSets(fixture.Namespace).List(metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())

		Expect(statefulsets.Items).To(HaveLen(1))
		Expect(statefulsets.Items[0].Name).To(ContainSubstring("the-app-guid"))
	})

	FIt("sets the cf instance index env var on each container", func() {
		Eventually(func() int32 {
			statefulsets, err := fixture.Clientset.AppsV1().StatefulSets(fixture.Namespace).List(metav1.ListOptions{})
			if err != nil {
				return -1
			}

			if len(statefulsets.Items) != 1 {
				return -1
			}
			return statefulsets.Items[0].Status.ReadyReplicas
		}, "10s").Should(Equal(int32(2)))

		pods, err := fixture.Clientset.CoreV1().Pods(fixture.Namespace).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(pods.Items).To(HaveLen(2))

		containers := pods.Items[0].Spec.Containers

		req := fixture.Clientset.CoreV1().Pods(fixture.Namespace).GetLogs(pods.Items[0].Name, &v1.PodLogOptions{})
		logs, err := req.Stream()
		Expect(err).NotTo(HaveOccurred())
		defer logs.Close()
		logBytes, err := ioutil.ReadAll(logs)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(logBytes)).To(Equal("jim"))

		instanceIds := []string{getEnvVar(containers[0], eirini.EnvCFInstanceIndex), getEnvVar(containers[1], eirini.EnvCFInstanceIndex)}
		Expect(instanceIds).To(ConsistOf("1", "2"))
	})

	Context("when the app has user defined annotations", func() {
		BeforeEach(func() {
			body = `{
			"guid": "the-app-guid",
			"version": "0.0.0",
			"ports" : [8080],
		  "lifecycle": {
				"docker_lifecycle": {
				  "image": "foo",
					"command": ["bar", "baz"]
				}
			},
			"user_defined_annotations": {
			  "prometheus.io/scrape": "yes, please"
			}
		}`
		})

		It("should set the annotations to the pod template", func() {
			statefulsets, err := fixture.Clientset.AppsV1().StatefulSets(fixture.Namespace).List(metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(statefulsets.Items[0].Spec.Template.Annotations).To(HaveKeyWithValue("prometheus.io/scrape", "yes, please"))
		})
	})
})

func getEnvVar(container v1.Container, varName string) string {
	for _, envVar := range container.Env {
		if envVar.Name == varName {
			return envVar.Value
		}
	}

	Fail(fmt.Sprintf("Variable %q not found in %v", varName, container.Env))
	return ""
}
