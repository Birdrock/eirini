// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "code.cloudfoundry.org/eirini/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// LRPs returns a LRPInformer.
	LRPs() LRPInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// LRPs returns a LRPInformer.
func (v *version) LRPs() LRPInformer {
	return &lRPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
