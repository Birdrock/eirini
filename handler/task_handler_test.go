package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/eirini/eirinifakes"
	. "code.cloudfoundry.org/eirini/handler"
	"code.cloudfoundry.org/eirini/models/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager/lagertest"
)

var _ = Describe("TaskHandler", func() {

	var (
		ts            *httptest.Server
		logger        *lagertest.TestLogger
		buildpackTask *eirinifakes.FakeBifrostTask

		response *http.Response
		body     string
		path     string
		method   string
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		buildpackTask = new(eirinifakes.FakeBifrostTask)

		method = "POST"
		path = "/tasks/guid_1234"
		body = `{
				"app_guid": "our-app-id",
				"environment": [{"name": "HOWARD", "value": "the alien"}],
				"completion_callback": "example.com/call/me/maybe",
				"lifecycle": {
          "buildpack_lifecycle": {
						"droplet_guid": "some-guid",
						"droplet_hash": "some-hash",
					  "start_command": "some command"
					}
				}
			}`
	})

	JustBeforeEach(func() {
		handler := New(nil, nil, nil, buildpackTask, logger)
		ts = httptest.NewServer(handler)
		req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader([]byte(body)))
		Expect(err).NotTo(HaveOccurred())

		client := &http.Client{}
		response, err = client.Do(req)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ts.Close()
	})

	It("should return 202 Accepted code", func() {
		Expect(response.StatusCode).To(Equal(http.StatusAccepted))
	})

	It("should transfer the task", func() {
		Expect(buildpackTask.TransferTaskCallCount()).To(Equal(1))
		_, actualTaskGUID, actualTaskRequest := buildpackTask.TransferTaskArgsForCall(0)
		Expect(actualTaskGUID).To(Equal("guid_1234"))
		Expect(actualTaskRequest).To(Equal(cf.TaskRequest{
			AppGUID:            "our-app-id",
			Environment:        []cf.EnvironmentVariable{{Name: "HOWARD", Value: "the alien"}},
			CompletionCallback: "example.com/call/me/maybe",
			Lifecycle: cf.Lifecycle{
				BuildpackLifecycle: &cf.BuildpackLifecycle{
					DropletGUID:  "some-guid",
					DropletHash:  "some-hash",
					StartCommand: "some command",
				},
			},
		}))
	})

	When("transferring the task fails", func() {
		BeforeEach(func() {
			buildpackTask.TransferTaskReturns(errors.New("transfer-task-err"))
		})

		It("should return 500 Internal Server Error code", func() {
			Expect(response.StatusCode).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("when the request body cannot be unmarshalled", func() {
		BeforeEach(func() {
			body = "random stuff"
		})

		It("should return 400 Bad Request code", func() {
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should not transfer the task", func() {
			Expect(buildpackTask.TransferTaskCallCount()).To(Equal(0))
		})
	})
})
