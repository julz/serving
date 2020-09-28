/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var domainMappingCondSet = apis.NewLivingConditionSet(
	DomainMappingConditionReady,
	DomainMappingConditionIngressReady,
)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*DomainMapping) GetConditionSet() apis.ConditionSet {
	return domainMappingCondSet
}

// GetGroupVersionKind returns the GroupVersionKind.
func (r *DomainMapping) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("DomainMapping")
}

// IsReady returns if the domain mapping is configured and the ingress is serving the requested hostname.
func (rs *DomainMappingStatus) IsReady() bool {
	return domainMappingCondSet.Manage(rs).IsHappy()
}

// IsReady returns true if the Status condition DomainMappingConditionReady
// is true and the latest spec has been observed.
func (r *DomainMapping) IsReady() bool {
	rs := r.Status
	return rs.ObservedGeneration == r.Generation &&
		rs.GetCondition(DomainMappingConditionReady).IsTrue()
}

// IsFailed returns true if the resource has observed
// the latest generation and ready is false.
func (r *DomainMapping) IsFailed() bool {
	rs := r.Status
	return rs.ObservedGeneration == r.Generation &&
		rs.GetCondition(DomainMappingConditionReady).IsFalse()
}

// InitializeConditions sets the initial values to the conditions.
func (rs *DomainMappingStatus) InitializeConditions() {
	domainMappingCondSet.Manage(rs).InitializeConditions()
}

// MarkIngressNotConfigured changes the IngressReady condition to be unknown to reflect
// that the Ingress does not yet have a Status
func (rs *DomainMappingStatus) MarkIngressNotConfigured() {
	domainMappingCondSet.Manage(rs).MarkUnknown(DomainMappingConditionIngressReady,
		"IngressNotConfigured", "Ingress has not yet been reconciled.")
}
