/*
Copyright 2019 the original author or authors.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *KafkaProvider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-streaming-projectriff-io-v1alpha1-kafkaprovider,mutating=true,failurePolicy=fail,groups=streaming.projectriff.io,resources=kafkaproviders,verbs=create;update,versions=v1alpha1,name=kafkaproviders.build.projectriff.io

var _ webhook.Defaulter = &KafkaProvider{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KafkaProvider) Default() {
	// TODO implement
}

// +kubebuilder:webhook:path=/validate-streaming-projectriff-io-v1alpha1-kafkaprovider,mutating=false,failurePolicy=fail,groups=streaming.projectriff.io,resources=kafkaproviders,verbs=create;update,versions=v1alpha1,name=kafkaproviders.build.projectriff.io

var _ webhook.Validator = &KafkaProvider{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaProvider) ValidateCreate() error {
	// TODO implement
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaProvider) ValidateUpdate(old runtime.Object) error {
	// TODO implement
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaProvider) ValidateDelete() error {
	// TODO implement
	return nil
}