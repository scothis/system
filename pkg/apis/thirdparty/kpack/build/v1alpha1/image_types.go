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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImageSpec is the spec for a Image resource.
type ImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Tag                      string                 `json:"tag"`
	Builder                  corev1.ObjectReference `json:"builder,omitempty"`
	ServiceAccount           string                 `json:"serviceAccount,omitempty"`
	Source                   SourceConfig           `json:"source"`
	CacheSize                *resource.Quantity     `json:"cacheSize,omitempty"`
	FailedBuildHistoryLimit  *int64                 `json:"failedBuildHistoryLimit,omitempty"`
	SuccessBuildHistoryLimit *int64                 `json:"successBuildHistoryLimit,omitempty"`
	ImageTaggingStrategy     ImageTaggingStrategy   `json:"imageTaggingStrategy,omitempty"`
	Build                    *ImageBuild            `json:"build,omitempty"`
}

type SourceConfig struct {
	Git      *Git      `json:"git,omitempty"`
	Blob     *Blob     `json:"blob,omitempty"`
	Registry *Registry `json:"registry,omitempty"`
	SubPath  string    `json:"subPath,omitempty"`
}

type Git struct {
	URL      string `json:"url"`
	Revision string `json:"revision"`
}

type Blob struct {
	URL string `json:"url"`
}

type Registry struct {
	Image            string                        `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
}

type ImageTaggingStrategy string

const (
	None        ImageTaggingStrategy = "None"
	BuildNumber ImageTaggingStrategy = "BuildNumber"
)

type ImageBuild struct {
	Bindings  Bindings                    `json:"bindings,omitempty"`
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Bindings []Binding

type Binding struct {
	Name        string                       `json:"name,omitempty"`
	MetadataRef *corev1.LocalObjectReference `json:"metadataRef,omitempty"`
	SecretRef   *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ImageStatus is the status for a Image resource
type ImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status                `json:",inline"`
	LatestBuildRef             string `json:"latestBuildRef,omitempty"`
	LatestBuildImageGeneration int64  `json:"latestBuildImageGeneration,omitempty"`
	LatestImage                string `json:"latestImage,omitempty"`
	LatestStack                string `json:"latestStack,omitempty"`
	BuildCounter               int64  `json:"buildCounter,omitempty"`
	BuildCacheName             string `json:"buildCacheName,omitempty"`
	LatestBuildReason          string `json:"latestBuildReason,omitempty"`
}

// +kubebuilder:object:root=true

type Image struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageList contains a list of Image
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Image{}, &ImageList{})
}
