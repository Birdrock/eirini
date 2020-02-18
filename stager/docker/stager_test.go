package docker_test

import (
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/stager/docker"
	"code.cloudfoundry.org/eirini/stager/docker/dockerfakes"
	"code.cloudfoundry.org/eirini/stager/stagerfakes"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var _ = Describe("DockerStager", func() {

	var (
		stager           docker.Stager
		fetcher          *dockerfakes.FakeImageMetadataFetcher
		stagingCompleter *stagerfakes.FakeStagingCompleter
	)

	Context("Stage a docker image", func() {

		var (
			stagingErr     error
			stagingRequest cf.StagingRequest
		)

		BeforeEach(func() {
			fetcher = new(dockerfakes.FakeImageMetadataFetcher)
			stagingCompleter = new(stagerfakes.FakeStagingCompleter)
			stagingRequest = cf.StagingRequest{
				CompletionCallback: "the-completion-callback/call/me",
				Lifecycle: cf.StagingLifecycle{
					DockerLifecycle: &cf.StagingDockerLifecycle{
						Image: "eirini/some-app:some-tag",
					},
				},
			}

			fetcher.Returns(&v1.ImageConfig{
				ExposedPorts: map[string]struct{}{
					"8888/tcp": {},
				},
			}, nil)
		})

		JustBeforeEach(func() {
			stager = docker.Stager{
				Logger:               lagertest.NewTestLogger(""),
				ImageMetadataFetcher: fetcher.Spy,
				StagingCompleter:     stagingCompleter,
			}

			stagingErr = stager.Stage("stg-guid", stagingRequest)
		})

		It("should succeed", func() {
			Expect(stagingErr).ToNot(HaveOccurred())
		})

		It("should create the correct docker image ref", func() {
			Expect(fetcher.CallCount()).To(Equal(1))
			ref, _ := fetcher.ArgsForCall(0)
			Expect(ref).To(Equal("//docker.io/eirini/some-app:some-tag"))
		})

		It("should complete staging with correct parameters", func() {
			Expect(stagingCompleter.CompleteStagingCallCount()).To(Equal(1))
			task := stagingCompleter.CompleteStagingArgsForCall(0)

			Expect(task.TaskGuid).To(Equal("stg-guid"))
			Expect(task.Failed).To(BeFalse())
			Expect(task.Annotation).To(Equal(`{"completion_callback": "the-completion-callback/call/me"}`))

			var payload docker.StagingCallbackPayload
			Expect(json.Unmarshal([]byte(task.Result), &payload)).To(Succeed())

			Expect(payload.Result.LifecycleType).To(Equal("docker"))
			Expect(payload.Result.LifecycleMetadata.DockerImage).To(Equal("eirini/some-app:some-tag"))
			Expect(payload.Result.ProcessTypes.Web).To(BeEmpty())
			Expect(payload.Result.ExecutionMetadata).To(Equal(`{"cmd":[],"ports":[{"Port":8888,"Protocol":"tcp"}]}`))
		})

		Context("when the image is from the standard library", func() {
			BeforeEach(func() {
				stagingRequest.Lifecycle.DockerLifecycle.Image = "ubuntu"
			})

			It("should succeed", func() {
				Expect(stagingErr).ToNot(HaveOccurred())
			})

			It("should create the correct docker image ref", func() {
				Expect(fetcher.CallCount()).To(Equal(1))
				ref, _ := fetcher.ArgsForCall(0)
				Expect(ref).To(Equal("//docker.io/library/ubuntu"))
			})
		})

		Context("when the image is from a private registry", func() {
			BeforeEach(func() {
				stagingRequest.Lifecycle.DockerLifecycle.Image = "private-registry.io/user/repo"
				stagingRequest.Lifecycle.DockerLifecycle.RegistryUsername = "some-user"
				stagingRequest.Lifecycle.DockerLifecycle.RegistryPassword = "thepasswrd"

			})

			It("should succeed", func() {
				Expect(stagingErr).ToNot(HaveOccurred())
			})

			It("should create the correct docker image ref", func() {
				Expect(fetcher.CallCount()).To(Equal(1))
				ref, _ := fetcher.ArgsForCall(0)
				Expect(ref).To(Equal("//private-registry.io/user/repo"))
			})

			It("should provide the correct credentials", func() {
				Expect(fetcher.CallCount()).To(Equal(1))
				_, ctx := fetcher.ArgsForCall(0)
				Expect(ctx.DockerAuthConfig.Username).To(Equal("some-user"))
				Expect(ctx.DockerAuthConfig.Password).To(Equal("thepasswrd"))
			})
		})

		Context("when the staging completion callback fails", func() {
			BeforeEach(func() {
				stagingCompleter.CompleteStagingReturns(errors.New("callback failed"))
			})

			It("should fail with the right error", func() {
				Expect(stagingErr).To(MatchError("callback failed"))
			})
		})

		Context("when the image ref is invalid", func() {
			BeforeEach(func() {
				stagingRequest.Lifecycle.DockerLifecycle.Image = "this is invalid"
			})

			It("should fail with the right error", func() {
				Expect(stagingErr).ToNot(HaveOccurred())
				Expect(stagingCompleter.CompleteStagingCallCount()).To(Equal(1))

				taskCallbackResponse := stagingCompleter.CompleteStagingArgsForCall(0)
				Expect(taskCallbackResponse.Failed).To(BeTrue())
				Expect(taskCallbackResponse.FailureReason).To(ContainSubstring("failed to parse image ref"))
			})

			Context("when the staging completion callback fails", func() {
				BeforeEach(func() {
					stagingCompleter.CompleteStagingReturns(errors.New("callback failed"))
				})

				It("should fail with the right error", func() {
					Expect(stagingErr).To(MatchError("callback failed"))
				})
			})
		})

		Context("when metadata fetching fails", func() {
			BeforeEach(func() {
				fetcher.Returns(nil, errors.New("boom"))
			})

			It("should fail with the right error", func() {
				Expect(stagingErr).ToNot(HaveOccurred())
				Expect(stagingCompleter.CompleteStagingCallCount()).To(Equal(1))

				taskCallbackResponse := stagingCompleter.CompleteStagingArgsForCall(0)
				Expect(taskCallbackResponse.Failed).To(BeTrue())
				Expect(taskCallbackResponse.FailureReason).To(ContainSubstring("failed to fetch image metadata"))
			})

			Context("when the staging completion callback fails", func() {
				BeforeEach(func() {
					stagingCompleter.CompleteStagingReturns(errors.New("callback failed"))
				})

				It("should fail with the right error", func() {
					Expect(stagingErr).To(MatchError("callback failed"))
				})
			})
		})

		Context("when exposed ports are wrongly formatted in the image metadata", func() {
			BeforeEach(func() {
				fetcher.Returns(&v1.ImageConfig{
					ExposedPorts: map[string]struct{}{
						"invalid-port-spec": {},
					},
				}, nil)
			})

			It("should respond to the callback url with failure", func() {
				Expect(stagingErr).ToNot(HaveOccurred())
				Expect(stagingCompleter.CompleteStagingCallCount()).To(Equal(1))

				taskCallbackResponse := stagingCompleter.CompleteStagingArgsForCall(0)
				Expect(taskCallbackResponse.Failed).To(BeTrue())
				Expect(taskCallbackResponse.FailureReason).To(ContainSubstring("failed to parse exposed ports"))
			})
		})

		Context("when the staging completion callback fails", func() {
			BeforeEach(func() {
				stagingCompleter.CompleteStagingReturns(errors.New("callback failed"))
			})

			It("should fail with the right error", func() {
				Expect(stagingErr).To(MatchError("callback failed"))
			})
		})
	})

})
