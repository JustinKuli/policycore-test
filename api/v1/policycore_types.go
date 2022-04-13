/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyCoreSpec struct {
	Severity          Severity          `json:"severity,omitempty"`
	RemediationAction RemediationAction `json:"remediationAction,omitempty"`
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`
}

//+kubebuilder:validation:Enum=low;medium;high;critical
type Severity string

const (
	LowSeverity      Severity = "low"
	MediumSeverity   Severity = "medium"
	HighSeverity     Severity = "high"
	CriticalSeverity Severity = "critical"
)

//+kubebuilder:validation:Enum=inform;enforce
type RemediationAction string

const (
	Inform  RemediationAction = "inform"
	Enforce RemediationAction = "enforce"
)

//+kubebuilder:validation:Required
type NamespaceSelector struct {
	Include []NonEmptyString `json:"include,omitempty"`
	Exclude []NonEmptyString `json:"exclude,omitempty"`
}

//+kubebuilder:validation:MinLength=1
type NonEmptyString string

type PolicyCoreStatus struct {
	ComplianceState ComplianceState `json:"compliant,omitempty"`
	RelatedObjects  []RelatedObject `json:"relatedObjects,omitempty"`
}

//+kubebuilder:validation:Enum=Compliant;NonCompliant;UnknownCompliancy
type ComplianceState string

const (
	// Compliant is an ComplianceState
	Compliant ComplianceState = "Compliant"

	// NonCompliant is an ComplianceState
	NonCompliant ComplianceState = "NonCompliant"

	// UnknownCompliancy is an ComplianceState
	UnknownCompliancy ComplianceState = "UnknownCompliancy"
)

type RelatedObject struct {
	Object    ObjectRef       `json:"object,omitempty"`
	Compliant ComplianceState `json:"compliant,omitempty"`
	Reason    string          `json:"reason,omitempty"`
}

type ObjectRef struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        ReferenceMetadata `json:"metadata,omitempty"`
}

type ReferenceMetadata struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PolicyCore is the Schema for the policycores API
type PolicyCore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyCoreSpec   `json:"spec,omitempty"`
	Status PolicyCoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PolicyCoreList contains a list of PolicyCore
type PolicyCoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyCore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyCore{}, &PolicyCoreList{})
}
