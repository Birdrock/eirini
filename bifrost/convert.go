package bifrost

import (
	"fmt"

	"github.com/pkg/errors"

	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/lager"
)

type DropletToImageConverter struct {
	logger     lager.Logger
	registryIP string
}

func NewConverter(logger lager.Logger, registryIP string) *DropletToImageConverter {
	return &DropletToImageConverter{
		logger:     logger,
		registryIP: registryIP,
	}
}

func (c *DropletToImageConverter) Convert(request cf.DesireLRPRequest) (opi.LRP, error) {
	vcapJSON := request.Environment["VCAP_APPLICATION"]
	vcap, err := parseVcapApplication(vcapJSON)
	if err != nil {
		return opi.LRP{}, errors.Wrap(err, "failed to parse vcap app")
	}

	var command []string
	var env map[string]string
	var image string

	env = map[string]string{
		"LANG":              "en_US.UTF-8",
		"CF_INSTANCE_ADDR":  "0.0.0.0:8080",
		"CF_INSTANCE_PORT":  "8080",
		"CF_INSTANCE_PORTS": `[{"external":8080,"internal":8080}]`,
	}

	healthcheck := opi.Healtcheck{
		Type:      request.HealthCheckType,
		Endpoint:  request.HealthCheckHTTPEndpoint,
		TimeoutMs: request.HealthCheckTimeoutMs,
	}

	switch {
	case request.Lifecycle.DockerLifecycle != nil:
		image = request.Lifecycle.DockerLifecycle.Image
		command = request.Lifecycle.DockerLifecycle.Command
		healthcheck.Port = request.Ports[0]

	case request.Lifecycle.BuildpackLifecycle != nil:
		var buildpackEnv map[string]string
		lifecycle := request.Lifecycle.BuildpackLifecycle
		image, command, buildpackEnv, healthcheck.Port = c.buildpackProperties(lifecycle.DropletGUID, lifecycle.DropletHash, lifecycle.StartCommand)
		env = mergeMaps(env, buildpackEnv)

	case request.DropletGUID != "":
		var buildpackEnv map[string]string
		image, command, buildpackEnv, healthcheck.Port = c.buildpackProperties(request.DropletGUID, request.DropletHash, request.StartCommand)
		env = mergeMaps(env, buildpackEnv)

	default:
		return opi.LRP{}, fmt.Errorf("missing lifecycle data")
	}

	routesJSON := getRequestedRoutes(request)

	identifier := opi.LRPIdentifier{
		GUID:    request.GUID,
		Version: request.Version,
	}

	volumeMounts := []opi.VolumeMount{}

	for _, vm := range request.VolumeMounts {
		volumeMounts = append(volumeMounts, opi.VolumeMount{
			MountPath: vm.MountDir,
			ClaimName: vm.VolumeID,
		})
	}

	return opi.LRP{
		AppName:         vcap.AppName,
		SpaceName:       vcap.SpaceName,
		LRPIdentifier:   identifier,
		Image:           image,
		TargetInstances: request.NumInstances,
		Command:         command,
		Env:             mergeMaps(request.Environment, env),
		Health:          healthcheck,
		Ports:           request.Ports,
		Metadata: map[string]string{
			cf.VcapAppName: vcap.AppName,
			cf.VcapAppID:   vcap.AppID,
			cf.VcapVersion: vcap.Version,
			cf.ProcessGUID: request.ProcessGUID,
			cf.VcapAppUris: routesJSON,
			cf.LastUpdated: request.LastUpdated,
		},
		MemoryMB:     request.MemoryMB,
		DiskMB:       request.DiskMB,
		CPUWeight:    request.CPUWeight,
		VolumeMounts: volumeMounts,
		LRP:          request.LRP,
	}, nil
}

func (c *DropletToImageConverter) buildpackProperties(dropletGUID, dropletHash, startCommand string) (string, []string, map[string]string, int32) {
	image := c.imageURI(dropletGUID, dropletHash)
	command := []string{"dumb-init", "--", "/lifecycle/launch"}
	buildpackEnv := map[string]string{
		"HOME":          "/home/vcap/app",
		"PATH":          "/usr/local/bin:/usr/bin:/bin",
		"USER":          "vcap",
		"TMPDIR":        "/home/vcap/tmp",
		"START_COMMAND": startCommand,
	}

	return image, command, buildpackEnv, int32(8080)
}

func getRequestedRoutes(request cf.DesireLRPRequest) string {
	routes := request.Routes
	if routes == nil {
		return ""
	}
	if _, ok := routes["cf-router"]; !ok {
		return ""
	}

	cfRouterRoutes := routes["cf-router"]
	data, err := cfRouterRoutes.MarshalJSON()
	if err != nil {
		panic("This should never happen!")
	}

	return string(data)
}

func (c *DropletToImageConverter) imageURI(dropletGUID, dropletHash string) string {
	return fmt.Sprintf("%s/cloudfoundry/%s:%s", c.registryIP, dropletGUID, dropletHash)
}

func mergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
