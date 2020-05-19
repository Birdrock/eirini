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

	waitTillSetReady := func() {
		EventuallyWithOffset(1, func() bool {
			statefulsets, err := fixture.Clientset.AppsV1().StatefulSets(fixture.Namespace).List(metav1.ListOptions{})
			if err != nil || len(statefulsets.Items) != 1 {
				return false
			}
			return statefulsets.Items[0].Status.ReadyReplicas == *statefulsets.Items[0].Spec.Replicas
		}, "10s").Should(BeTrue())
	}

	logsForPod := func(idx int) string {
		pods := fixture.Clientset.CoreV1().Pods(fixture.Namespace)
		podList, err := pods.List(metav1.ListOptions{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		logs, err := pods.GetLogs(podList.Items[idx].Name, &v1.PodLogOptions{}).Stream()
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		defer logs.Close()
		logBytes, err := ioutil.ReadAll(logs)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return string(logBytes)
	}

	FIt("sets the CF_INSTANCE_INDEX env var on each container", func() {
		waitTillSetReady()
		for i := 0; i < 2; i++ {
			logs := logsForPod(i)
			Expect(logs).To(ContainSubstring(fmt.Sprintf("%s=%d\n", eirini.EnvCFInstanceIndex, i)))
		}
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
