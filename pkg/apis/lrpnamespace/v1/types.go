package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LrpNamespace describes an LrpNamespace resource
type LrpNamespace struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec LrpNamespaceSpec `json:"spec"`
}

// LrpNamespaceSpec is the spec for a LrpNamespace resource
type LrpNamespaceSpec struct {
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LrpNamespaceList is a list of LrpNamespace resources
type LrpNamespaceList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []LrpNamespace `json:"items"`
}
