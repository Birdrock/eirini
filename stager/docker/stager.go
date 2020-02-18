package docker

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/stager"
	"code.cloudfoundry.org/lager"
	"github.com/containers/image/types"
	"github.com/docker/distribution/reference"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . ImageMetadataFetcher
type ImageMetadataFetcher func(string, types.SystemContext) (*v1.ImageConfig, error)

func (f ImageMetadataFetcher) Fetch(dockerRef string, sysCtx types.SystemContext) (*v1.ImageConfig, error) {
	return f(dockerRef, sysCtx)
}

type Stager struct {
	Logger               lager.Logger
	ImageMetadataFetcher ImageMetadataFetcher
	StagingCompleter     stager.StagingCompleter
}

type StagingCallbackPayload struct {
	Result StagingResult `json:"result"`
}

type StagingResult struct {
	LifecycleType     string            `json:"lifecycle_type"`
	LifecycleMetadata LifecycleMetadata `json:"lifecycle_metadata"`
	ProcessTypes      ProcessTypes      `json:"process_types"`
	ExecutionMetadata string            `json:"execution_metadata"`
}

type LifecycleMetadata struct {
	DockerImage string `json:"docker_image"`
}

type ProcessTypes struct {
	Web string `json:"web"`
}

type port struct {
	Port     uint   `json:"Port"`
	Protocol string `json:"Protocol"`
}

type executionMetadata struct {
	Cmd   []string `json:"cmd"`
	Ports []port   `json:"ports"`
}

func (s Stager) Stage(stagingGUID string, request cf.StagingRequest) error {
	taskCallbackResponse := &models.TaskCallbackResponse{
		TaskGuid:   stagingGUID,
		Annotation: fmt.Sprintf(`{"completion_callback": "%s"}`, request.CompletionCallback),
	}

	imageConfig, err := s.getImageConfig(request.Lifecycle.DockerLifecycle)
	if err != nil {
		s.Logger.Error("failed to get image config", err)
		return s.respondWithFailure(taskCallbackResponse, errors.Wrap(err, "failed to get image config"))
	}

	ports, err := parseExposedPorts(imageConfig)
	if err != nil {
		s.Logger.Error("failed to parse exposed ports", err)
		return s.respondWithFailure(taskCallbackResponse, errors.Wrap(err, "failed to parse exposed ports"))
	}

	stagingResult, err := buildStagingResult(request.Lifecycle.DockerLifecycle.Image, ports)
	if err != nil {
		s.Logger.Error("failed to build staging result", err)
		return s.respondWithFailure(taskCallbackResponse, errors.Wrap(err, "failed to build staging result"))
	}

	taskCallbackResponse.Result = stagingResult
	return s.CompleteStaging(taskCallbackResponse)
}

func (s Stager) respondWithFailure(taskCallbackResponse *models.TaskCallbackResponse, err error) error {
	taskCallbackResponse.Failed = true
	taskCallbackResponse.FailureReason = err.Error()
	return s.CompleteStaging(taskCallbackResponse)
}

func (s Stager) CompleteStaging(task *models.TaskCallbackResponse) error {
	return s.StagingCompleter.CompleteStaging(task)
}

func (s Stager) getImageConfig(lifecycle *cf.StagingDockerLifecycle) (*v1.ImageConfig, error) {
	named, err := reference.ParseNormalizedNamed(lifecycle.Image)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse image ref")
	}
	dockerRef := fmt.Sprintf("//%s", named.String())
	imgMetadata, err := s.ImageMetadataFetcher.Fetch(dockerRef, types.SystemContext{
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: lifecycle.RegistryUsername,
			Password: lifecycle.RegistryPassword,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch image metadata")
	}

	return imgMetadata, nil
}

func parseExposedPorts(imageConfig *v1.ImageConfig) ([]port, error) {
	var (
		portNum  uint
		protocol string
	)

	ports := make([]port, 0, len(imageConfig.ExposedPorts))
	for imagePort := range imageConfig.ExposedPorts {
		_, err := fmt.Sscanf(imagePort, "%d/%s", &portNum, &protocol)
		if err != nil {
			return []port{}, err
		}
		ports = append(ports, port{
			Port:     portNum,
			Protocol: protocol,
		})
	}
	return ports, nil
}

func buildStagingResult(image string, ports []port) (string, error) {
	executionMetadataJSON, err := json.Marshal(executionMetadata{
		Cmd:   []string{},
		Ports: ports,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to parse execution metadata")
	}

	payload := StagingCallbackPayload{
		Result: StagingResult{
			LifecycleType: "docker",
			LifecycleMetadata: LifecycleMetadata{
				DockerImage: image,
			},
			ProcessTypes:      ProcessTypes{Web: ""},
			ExecutionMetadata: string(executionMetadataJSON),
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Wrap(err, "failed to build payload json")
	}

	return string(payloadJSON), nil
}
