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

// KafkaTopicSpec defines the desired state of a Kafka topic.
type KafkaTopicSpec struct {
	// Partitions is the number of partitions for the topic.
	// +kubebuilder:validation:Minimum=1
	Partitions int32 `json:"partitions"`

	// Replicas is the replication factor for the topic.
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic.
type KafkaTopicStatus struct {
	// Partitions is the number of partitions in the topic.
	// +optional
	Partitions int32 `json:"partitions,omitempty"`

	// Replicas is the replication factor of the topic.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Conditions provides information about the topic's current state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KafkaTopic is the Schema for the kafkatopics API
type KafkaTopic struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of KafkaTopic
	// +required
	Spec KafkaTopicSpec `json:"spec"`

	// status defines the observed state of KafkaTopic
	// +optional
	Status KafkaTopicStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// KafkaTopicList contains a list of KafkaTopic
type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
