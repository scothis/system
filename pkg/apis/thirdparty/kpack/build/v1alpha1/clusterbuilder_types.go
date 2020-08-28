/*
Copyright 2020 The original author or authors

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
	"github.com/vmware-labs/reconciler-runtime/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterBuilderSpec is the spec for a ClusterBuilder resource.
type ClusterBuilderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	BuilderSpec       `json:",inline"`
	ServiceAccountRef corev1.ObjectReference `json:"serviceAccountRef,omitempty"`
}

type BuilderSpec struct {
	Tag   string                 `json:"tag,omitempty"`
	Stack corev1.ObjectReference `json:"stack,omitempty"`
	Store corev1.ObjectReference `json:"store,omitempty"`
	Order []OrderEntry           `json:"order,omitempty"`
}

type OrderEntry struct {
	Group []BuildpackRef `json:"group,omitempty"`
}

type BuildpackRef struct {
	BuildpackInfo `json:",inline"`
	Optional      bool `json:"optional,omitempty"`
}

type BuildpackInfo struct {
	Id      string `json:"id"`
	Version string `json:"version,omitempty"`
}

// ClusterBuilderStatus is the status for a ClusterBuilder resource
type ClusterBuilderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status     `json:",inline"`
	BuilderMetadata BuildpackMetadataList `json:"builderMetadata,omitempty"`
	Stack           BuildStack            `json:"stack,omitempty"`
	LatestImage     string                `json:"latestImage,omitempty"`
}

type BuildpackMetadataList []BuildpackMetadata

type BuildpackMetadata struct {
	Id       string `json:"id"`
	Version  string `json:"version"`
	Homepage string `json:"homepage,omitempty"`
}

type BuildStack struct {
	RunImage string `json:"runImage,omitempty"`
	ID       string `json:"id,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterBuilderSpec   `json:"spec,omitempty"`
	Status ClusterBuilderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterBuilderList contains a list of ClusterBuilder
type ClusterBuilderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBuilder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBuilder{}, &ClusterBuilderList{})
}
