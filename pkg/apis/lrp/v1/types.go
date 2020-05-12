package v1

import (
	"encoding/json"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MyResource describes a MyResource resource
type LRP struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	meta_v1.TypeMeta `json:",inline"`
	// ObjectMeta contains the metadata for the particular object, including
	// things like...
	//  - name
	//  - namespace
	//  - self link
	//  - labels
	//  - ... etc ...
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec LRPSpec `json:"spec"`
}

// MyResourceSpec is the spec for a MyResource resource
type LRPSpec struct {
	GUID                    string                     `json:"guid"`
	Version                 string                     `json:"version"`
	ProcessGUID             string                     `json:"process_guid"`
	ProcessType             string                     `json:"process_type"`
	AppGUID                 string                     `json:"app_guid"`
	AppName                 string                     `json:"app_name"`
	SpaceGUID               string                     `json:"space_guid"`
	SpaceName               string                     `json:"space_name"`
	OrganizationGUID        string                     `json:"organization_guid"`
	OrganizationName        string                     `json:"organization_name"`
	PlacementTags           []string                   `json:"placement_tags"`
	Ports                   []int32                    `json:"ports"`
	Routes                  map[string]json.RawMessage `json:"routes"`
	Environment             map[string]string          `json:"environment"`
	EgressRules             []json.RawMessage          `json:"egress_rules"`
	NumInstances            *int                       `json:"instances"`
	LastUpdated             string                     `json:"last_updated"`
	HealthCheckType         string                     `json:"health_check_type"`
	HealthCheckHTTPEndpoint string                     `json:"health_check_http_endpoint"`
	HealthCheckTimeoutMs    *uint                      `json:"health_check_timeout_ms"`
	StartTimeoutMs          *uint                      `json:"start_timeout_ms"`
	MemoryMB                *int64                     `json:"memory_mb"`
	DiskMB                  *int64                     `json:"disk_mb"`
	CPUWeight               *uint8                     `json:"cpu_weight"`
	VolumeMounts            []VolumeMount              `json:"volume_mounts"`
	Lifecycle               Lifecycle                  `json:"lifecycle"`
	DropletHash             string                     `json:"droplet_hash"`
	DropletGUID             string                     `json:"droplet_guid"`
	StartCommand            string                     `json:"start_command"`
	UserDefinedAnnotations  map[string]string          `json:"user_defined_annotations"`
	LRP                     string
}

type Lifecycle struct {
	DockerLifecycle    *DockerLifecycle    `json:"docker_lifecycle"`
	BuildpackLifecycle *BuildpackLifecycle `json:"buildpack_lifecycle"`
}

type DockerLifecycle struct {
	Image            string   `json:"image"`
	Command          []string `json:"command"`
	RegistryUsername string   `json:"registry_username"`
	RegistryPassword string   `json:"registry_password"`
}

type BuildpackLifecycle struct {
	DropletHash  string `json:"droplet_hash"`
	DropletGUID  string `json:"droplet_guid"`
	StartCommand string `json:"start_command"`
}

type VolumeMount struct {
	VolumeID string `json:"volume_id"`
	MountDir string `json:"mount_dir"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MyResourceList is a list of MyResource resources
type LRPList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []LRP `json:"items"`
}
