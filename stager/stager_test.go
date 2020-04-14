package stager_test

import (
	"errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi/opifakes"
	. "code.cloudfoundry.org/eirini/stager"
	"code.cloudfoundry.org/eirini/stager/stagerfakes"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stager", func() {

	var (
		stager           eirini.Stager
		taskDesirer      *opifakes.FakeTaskDesirer
		stagingCompleter *stagerfakes.FakeStagingCompleter
		err              error
	)

	BeforeEach(func() {
		taskDesirer = new(opifakes.FakeTaskDesirer)

		logger := lagertest.NewTestLogger("test")

		stagingCompleter = new(stagerfakes.FakeStagingCompleter)

		stager = &Stager{
			Desirer:          taskDesirer,
			StagingCompleter: stagingCompleter,
			Logger:           logger,
		}
	})

	Context("When staging", func() {
		It("returns not supported error", func() {
			Expect(stager.Stage("aaa", cf.StagingRequest{})).NotTo(Succeed())
		})
	})

	Context("When completing staging", func() {

		var (
			task *models.TaskCallbackResponse
		)

		BeforeEach(func() {
			annotation := `{"completion_callback": "some-cc-endpoint.io/call/me/maybe"}`

			task = &models.TaskCallbackResponse{
				TaskGuid:      "our-task-guid",
				Failed:        false,
				FailureReason: "",
				Result:        `{"very": "good"}`,
				Annotation:    annotation,
				CreatedAt:     123456123,
			}
		})

		JustBeforeEach(func() {
			err = stager.CompleteStaging(task)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should complete staging", func() {
			Expect(stagingCompleter.CompleteStagingCallCount()).To(Equal(1))
			Expect(stagingCompleter.CompleteStagingArgsForCall(0)).To(Equal(task))
		})

		It("should delete the task", func() {
			Expect(taskDesirer.DeleteCallCount()).To(Equal(1))

			taskName := taskDesirer.DeleteArgsForCall(0)
			Expect(taskName).To(Equal(task.TaskGuid))
		})

		Context("and the staging completer fails", func() {
			BeforeEach(func() {
				stagingCompleter.CompleteStagingReturns(errors.New("complete boom"))
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("complete boom"))
			})

			It("should delete the task", func() {
				Expect(taskDesirer.DeleteCallCount()).To(Equal(1))

				taskName := taskDesirer.DeleteArgsForCall(0)
				Expect(taskName).To(Equal(task.TaskGuid))
			})
		})

		Context("and the task deletion fails", func() {
			BeforeEach(func() {
				taskDesirer.DeleteReturns(errors.New("delete boom"))
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("delete boom"))
			})
		})
	})
})
