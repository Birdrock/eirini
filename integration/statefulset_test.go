package statefulsets_test

import (
	"fmt"

	"code.cloudfoundry.org/eirini/integration/util"
	. "code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/eirini/k8s/utils"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("StatefulSet Manager", func() {

	var (
		desirer opi.Desirer
		odinLRP *opi.LRP
		thorLRP *opi.LRP
	)

	BeforeEach(func() {
		odinLRP = createLRP("ödin")
		thorLRP = createLRP("thor")
	})

	AfterEach(func() {
		cleanupStatefulSet(odinLRP)
		cleanupStatefulSet(thorLRP)
		Eventually(func() []appsv1.StatefulSet {
			return listAllStatefulSets(odinLRP, thorLRP)
		}, timeout).Should(BeEmpty())
	})

	JustBeforeEach(func() {
		logger := lagertest.NewTestLogger("test")
		desirer = NewStatefulSetDesirer(
			clientset,
			namespace,
			"registry-secret",
			"rootfsversion",
			"default",
			"default",
			logger,
		)
	})

	Context("When creating a LRP", func() {

		JustBeforeEach(func() {
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())
			err = desirer.Desire(thorLRP)
			Expect(err).ToNot(HaveOccurred())
		})

		// join all tests in a single with By()
		It("should create a StatefulSet object", func() {
			statefulset := getStatefulSet(odinLRP)
			Expect(statefulset.Name).To(ContainSubstring(odinLRP.GUID))
			Expect(statefulset.Spec.Template.Spec.Containers[0].Command).To(Equal(odinLRP.Command))
			Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal(odinLRP.Image))
			Expect(statefulset.Spec.Replicas).To(Equal(int32ptr(odinLRP.TargetInstances)))
			Expect(statefulset.Annotations[AnnotationOriginalRequest]).To(Equal(odinLRP.LRP))
		})

		It("should create all associated pods", func() {
			var podNames []string

			Eventually(func() []string {
				podNames = podNamesFromPods(listPods(odinLRP.LRPIdentifier))
				return podNames
			}, timeout).Should(HaveLen(odinLRP.TargetInstances))

			for i := 0; i < odinLRP.TargetInstances; i++ {
				podIndex := i
				Expect(podNames[podIndex]).To(ContainSubstring(odinLRP.GUID))

				Eventually(func() corev1.PodPhase {
					return getPodPhase(podIndex, odinLRP.LRPIdentifier)
				}, timeout).Should(Equal(corev1.PodRunning))
			}

			statefulset := getStatefulSet(odinLRP)
			Expect(statefulset.Status.ReadyReplicas).To(Equal(statefulset.Status.Replicas))
		})

		It("should create a pod disruption budget for the lrp", func() {
			statefulset := getStatefulSet(odinLRP)
			pdb, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(pdb).NotTo(BeNil())
		})

		Context("when the lrp has 1 instance", func() {
			BeforeEach(func() {
				odinLRP.TargetInstances = 1
			})

			It("should not create a pod disruption budget for the lrp", func() {
				statefulset := getStatefulSet(odinLRP)
				_, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})

		Context("when additional app info is provided", func() {
			BeforeEach(func() {
				odinLRP.OrgName = "odin-org"
				odinLRP.OrgGUID = "odin-org-guid"
				odinLRP.SpaceName = "odin-space"
				odinLRP.SpaceGUID = "odin-space-guid"
			})

			DescribeTable("sets appropriate annotations to statefulset", func(key, value string) {
				statefulset := getStatefulSet(odinLRP)
				Expect(statefulset.Annotations).To(HaveKeyWithValue(key, value))
			},
				Entry("SpaceName", AnnotationSpaceName, "odin-space"),
				Entry("SpaceGUID", AnnotationSpaceGUID, "odin-space-guid"),
				Entry("OrgName", AnnotationOrgName, "odin-org"),
				Entry("OrgGUID", AnnotationOrgGUID, "odin-org-guid"),
			)

			It("sets appropriate labels to statefulset", func() {
				statefulset := getStatefulSet(odinLRP)
				Expect(statefulset.Labels).To(HaveKeyWithValue(LabelGUID, odinLRP.LRPIdentifier.GUID))
				Expect(statefulset.Labels).To(HaveKeyWithValue(LabelVersion, odinLRP.LRPIdentifier.Version))
				Expect(statefulset.Labels).To(HaveKeyWithValue(LabelSourceType, "APP"))
			})
		})

		Context("when the app has more than one instances", func() {
			BeforeEach(func() {
				odinLRP.TargetInstances = 2
			})

			It("should schedule app pods on different nodes", func() {
				if getNodeCount() == 1 {
					Skip("target cluster has only one node")
				}

				Eventually(func() []corev1.Pod {
					return listPods(odinLRP.LRPIdentifier)
				}, timeout).Should(HaveLen(2))

				nodeNames := nodeNamesFromPods(listPods(odinLRP.LRPIdentifier))

				Expect(nodeNames).To(HaveLen(2))
				Expect(nodeNames[0]).ToNot(Equal(nodeNames[1]))
			})
		})

		Context("When private docker registry credentials are provided", func() {
			BeforeEach(func() {
				odinLRP.Image = "eiriniuser/notdora:latest"
				odinLRP.PrivateRegistry = &opi.PrivateRegistry{
					Server:   "index.docker.io/v1/",
					Username: "eiriniuser",
					Password: util.GetEiriniDockerHubPassword(),
				}
			})

			It("creates a private registry secret", func() {
				statefulset := getStatefulSet(odinLRP)
				secret, err := getSecret(privateRegistrySecretName(statefulset.Name))
				Expect(err).NotTo(HaveOccurred())
				Expect(secret).NotTo(BeNil())
			})

			It("sets the ImagePullSecret correctly in the pod template", func() {
				Eventually(func() []corev1.Pod {
					return listPods(odinLRP.LRPIdentifier)
				}, timeout).Should(HaveLen(odinLRP.TargetInstances))

				for i := 0; i < odinLRP.TargetInstances; i++ {
					podIndex := i
					Eventually(func() corev1.PodPhase {
						return getPodPhase(podIndex, odinLRP.LRPIdentifier)
					}, timeout).Should(Equal(corev1.PodRunning))
				}
			})
		})

		Context("when we create the same StatefulSet again", func() {
			It("should error", func() {
				err := desirer.Desire(odinLRP)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When using a docker image that needs root access", func() {
			BeforeEach(func() {
				odinLRP.Image = "xanderstrike/web:astandke"
				odinLRP.Health.Type = "http"
				odinLRP.Health.Port = 8080
			})

			FIt("should start all the pods", func() {
				var podNames []string

				Eventually(func() []string {
					podNames = podNamesFromPods(listPods(odinLRP.LRPIdentifier))
					return podNames
				}, timeout).Should(HaveLen(odinLRP.TargetInstances))

				for i := 0; i < odinLRP.TargetInstances; i++ {
					podIndex := i
					Expect(podNames[podIndex]).To(ContainSubstring(odinLRP.GUID))

					Eventually(func() string {
						// TODO: rewrite me in a way that I can understand the output
						return utils.GetPodState(listPods(odinLRP.LRPIdentifier)[i])
					}, timeout).Should(Equal(opi.RunningState))
				}
				statefulset := getStatefulSet(odinLRP)

				Expect(statefulset.Status.ReadyReplicas).To(Equal(statefulset.Status.Replicas))
			})
		})
	})

	Context("When stopping a LRP", func() {
		var statefulsetName string

		JustBeforeEach(func() {
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() []corev1.Pod {
				return listPods(odinLRP.LRPIdentifier)
			}, timeout).Should(HaveLen(odinLRP.TargetInstances))

			statefulsetName = getStatefulSet(odinLRP).Name

			err = desirer.Stop(odinLRP.LRPIdentifier)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete the StatefulSet object", func() {
			Eventually(func() []appsv1.StatefulSet {
				return listStatefulSets("odin")
			}, timeout).Should(BeEmpty())
		})

		It("should delete the associated pods", func() {
			Eventually(func() []corev1.Pod {
				return listPods(odinLRP.LRPIdentifier)
			}, timeout).Should(BeEmpty())
		})

		It("should delete the pod disruption budget for the lrp", func() {
			Eventually(func() error {
				_, err := podDisruptionBudgets().Get(statefulsetName, v1.GetOptions{})
				return err
			}, timeout).Should(MatchError(ContainSubstring("not found")))
		})

		Context("when the lrp has only 1 instance", func() {
			BeforeEach(func() {
				odinLRP.TargetInstances = 1
			})

			It("keep the lrp without a pod disruption budget", func() {
				Eventually(func() error {
					_, err := podDisruptionBudgets().Get(statefulsetName, v1.GetOptions{})
					return err
				}, timeout).Should(MatchError(ContainSubstring("not found")))
			})
		})

		Context("When private docker registry credentials are provided", func() {
			BeforeEach(func() {
				odinLRP.Image = "eiriniuser/notdora:latest"
				odinLRP.PrivateRegistry = &opi.PrivateRegistry{
					Server:   "index.docker.io/v1/",
					Username: "eiriniuser",
					Password: util.GetEiriniDockerHubPassword(),
				}
			})

			It("should delete the StatefulSet object", func() {
				Eventually(func() []appsv1.StatefulSet {
					return listStatefulSets("odin")
				}, timeout).Should(BeEmpty())
			})

			It("should delete the private registry secret", func() {
				_, err := getSecret(privateRegistrySecretName(statefulsetName))
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})
	})

	Context("When updating a LRP", func() {
		var (
			instancesBefore int
			instancesAfter  int
		)

		JustBeforeEach(func() {
			odinLRP.TargetInstances = instancesBefore
			Expect(desirer.Desire(odinLRP)).To(Succeed())

			odinLRP.TargetInstances = instancesAfter
			Expect(desirer.Update(odinLRP)).To(Succeed())
		})

		Context("when scaling up from 1 to 2 instances", func() {
			BeforeEach(func() {
				instancesBefore = 1
				instancesAfter = 2
			})

			It("should create a pod disruption budget for the lrp", func() {
				statefulset := getStatefulSet(odinLRP)
				pdb, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(pdb).NotTo(BeNil())
			})
		})

		Context("when scaling up from 2 to 3 instances", func() {
			BeforeEach(func() {
				instancesBefore = 2
				instancesAfter = 3
			})

			It("should keep the existing pod disruption budget for the lrp", func() {
				statefulset := getStatefulSet(odinLRP)
				pdb, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(pdb).NotTo(BeNil())
			})
		})

		Context("when scaling down from 2 to 1 instances", func() {
			BeforeEach(func() {
				instancesBefore = 2
				instancesAfter = 1
			})

			It("should delete the pod disruption budget for the lrp", func() {
				statefulset := getStatefulSet(odinLRP)
				_, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})

		Context("when scaling down from 1 to 0 instances", func() {
			BeforeEach(func() {
				instancesBefore = 1
				instancesAfter = 0
			})

			It("should keep the lrp without a pod disruption budget", func() {
				statefulset := getStatefulSet(odinLRP)
				_, err := podDisruptionBudgets().Get(statefulset.Name, v1.GetOptions{})
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})

	})

	Context("When getting a LRP", func() {
		numberOfInstancesFn := func() int {
			lrp, err := desirer.Get(odinLRP.LRPIdentifier)
			Expect(err).ToNot(HaveOccurred())
			return lrp.RunningInstances
		}

		JustBeforeEach(func() {
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())
		})

		It("correctly reports the running instances", func() {
			Eventually(numberOfInstancesFn, timeout).Should(Equal(odinLRP.TargetInstances))
			Consistently(numberOfInstancesFn, "10s").Should(Equal(odinLRP.TargetInstances))
		})

		Context("When one of the instances if failing", func() {
			BeforeEach(func() {
				odinLRP = createLRP("odin")
				odinLRP.Health = opi.Healtcheck{
					Type: "port",
					Port: 3000,
				}
				odinLRP.Command = []string{
					"/bin/sh",
					"-c",
					`if [ $(echo $HOSTNAME | sed 's|.*-\(.*\)|\1|') -eq 1 ]; then
	exit;
else
	while true; do
		nc -lk -p 3000 -e echo just a server;
	done;
fi;`,
				}
			})

			It("correctly reports the running instances", func() {
				Eventually(numberOfInstancesFn, timeout).Should(Equal(1), fmt.Sprintf("pod %#v did not start", odinLRP.LRPIdentifier))
				Consistently(numberOfInstancesFn, "10s").Should(Equal(1), fmt.Sprintf("pod %#v did not keep running", odinLRP.LRPIdentifier))
			})
		})
	})

})

func int32ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func getPodPhase(index int, id opi.LRPIdentifier) corev1.PodPhase {
	pods := listPods(id)
	return pods[index].Status.Phase
}

func privateRegistrySecretName(statefulsetName string) string {
	return fmt.Sprintf("%s-registry-credentials", statefulsetName)
}
