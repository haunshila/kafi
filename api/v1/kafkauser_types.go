/*
Copyright 2025.

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

// KafkaUserSpec defines the desired state of a KafkaUser.
type KafkaUserSpec struct {
	// Authentication specifies the user's authentication method.
	// +kubebuilder:validation:Required
	Authentication KafkaUserAuthenticationSpec `json:"authentication"`

	// Authorization defines the ACLs for the user.
	// +optional
	Authorization KafkaUserAuthorizationSpec `json:"authorization,omitempty"`
}

// KafkaUserAuthenticationSpec defines the user authentication method.
// +kubebuilder:validation:Enum=plain;scram-sha-512
type KafkaUserAuthenticationType string

type KafkaUserAuthenticationSpec struct {
	Type KafkaUserAuthenticationType `json:"type"`
}

// KafkaUserAuthorizationSpec defines the user's ACLs.
type KafkaUserAuthorizationSpec struct {
	Acls []KafkaUserAcl `json:"acls"`
}

// KafkaUserAcl defines an access control list entry.
type KafkaUserAcl struct {
	// Topic is the topic the ACL applies to.
	Topic string `json:"topic"`
	// Operation is the allowed operation (e.g., READ, WRITE).
	// +kubebuilder:validation:Enum=read;write;all
	Operation string `json:"operation"`
}

// KafkaUserStatus defines the observed state of KafkaUser.
type KafkaUserStatus struct {
	// Conditions provides information about the user's current state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// SecretRef is a reference to the Secret containing the user's credentials.
	// +optional
	SecretRef string `json:"secretRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KafkaUser is the Schema for the kafkausers API
type KafkaUser struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of KafkaUser
	// +required
	Spec KafkaUserSpec `json:"spec"`

	// status defines the observed state of KafkaUser
	// +optional
	Status KafkaUserStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// KafkaUserList contains a list of KafkaUser
type KafkaUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaUser{}, &KafkaUserList{})
}
