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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaSpec defines the desired state of a Kafka cluster.
type KafkaSpec struct {
	// Replicas is the number of Kafka broker pods to run.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=3
	Replicas int32 `json:"replicas"`

	// Version is the Kafka version to use.
	// +kubebuilder:default:="3.5.0"
	Version string `json:"version"`

	// Storage defines the storage configuration for the Kafka brokers.
	// +kubebuilder:validation:Required
	Storage KafkaStorageSpec `json:"storage"`

	// Listeners specifies the Kafka listeners.
	Listeners []KafkaListenerSpec `json:"listeners"`

	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// KafkaStatus defines the observed state of Kafka.
type KafkaStatus struct {
	// Replicas is the number of ready Kafka broker pods.
	Replicas int32 `json:"replicas"`

	// Conditions provides information about the cluster's current state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// BootstrapServers is the address to connect to the Kafka cluster.
	BootstrapServers string `json:"bootstrapServers"`
}

// +kubebuilder:validation:Enum=persistent-claim;ephemeral
type KafkaStorageType string

// KafkaStorageSpec defines the storage configuration.
type KafkaStorageSpec struct {
	// StorageClassName is the name of the StorageClass to use for dynamic provisioning.
	// This StorageClass must be configured to use the Ceph CSI driver.
	StorageClassName string `json:"storageClassName"`

	// Size is the requested size of the persistent volume.
	// For example: "10Gi".
	Size string `json:"size"`
}

// KafkaListenerSpec defines a Kafka listener.
// +kubebuilder:validation:Enum=plain;tls
type KafkaListenerType string

type KafkaListenerSpec struct {
	Type KafkaListenerType `json:"type"`
	// Port is the port number for the listener.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Kafka is the Schema for the kafkas API
type Kafka struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Kafka
	// +required
	Spec KafkaSpec `json:"spec"`

	// status defines the observed state of Kafka
	// +optional
	Status KafkaStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// KafkaList contains a list of Kafka
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
